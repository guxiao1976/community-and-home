package manager

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	"github.com/guxiao/community-and-home/services/ai-model/rpc/model"
)

type CostManager struct {
	db                   sqlx.SqlConn
	callLogModel         model.AmCallLogModel
	usageStatisticsModel model.AmUsageStatisticsModel
	costAlertConfigModel model.AmCostAlertConfigModel
	alertRecordModel     model.AmAlertRecordModel
}

func NewCostManager(db sqlx.SqlConn, cacheConf cache.CacheConf) *CostManager {
	return &CostManager{
		db:                   db,
		callLogModel:         model.NewAmCallLogModel(db, cacheConf),
		usageStatisticsModel: model.NewAmUsageStatisticsModel(db, cacheConf),
		costAlertConfigModel: model.NewAmCostAlertConfigModel(db, cacheConf),
		alertRecordModel:     model.NewAmAlertRecordModel(db, cacheConf),
	}
}

func (m *CostManager) RecordCall(ctx context.Context, callLog *model.AmCallLog) error {
	_, err := m.callLogModel.Insert(ctx, callLog)
	if err != nil {
		return fmt.Errorf("insert call log: %w", err)
	}

	if err := m.updateStatistics(ctx, callLog); err != nil {
		logx.Errorf("failed to update statistics: %v", err)
	}

	if err := m.checkCostAlerts(ctx, callLog); err != nil {
		logx.Errorf("failed to check cost alerts: %v", err)
	}

	return nil
}

func (m *CostManager) updateStatistics(ctx context.Context, callLog *model.AmCallLog) error {
	date := callLog.CreatedTime.Format("2006-01-02")

	query := `
		INSERT INTO am_usage_statistics
		(model_id, model_name, stat_date, total_calls, success_calls, failed_calls, input_tokens, output_tokens, total_tokens, total_cost, avg_latency_ms)
		VALUES (?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			total_calls = total_calls + 1,
			success_calls = success_calls + ?,
			failed_calls = failed_calls + ?,
			input_tokens = input_tokens + ?,
			output_tokens = output_tokens + ?,
			total_tokens = total_tokens + ?,
			total_cost = total_cost + ?,
			avg_latency_ms = (avg_latency_ms * total_calls + ?) / (total_calls + 1)
	`

	successCount := int64(0)
	failedCount := int64(0)
	if callLog.Success == 1 {
		successCount = 1
	} else {
		failedCount = 1
	}

	latencyMs := int64(0)
	if callLog.LatencyMs.Valid {
		latencyMs = callLog.LatencyMs.Int64
	}

	totalTokens := callLog.InputTokens + callLog.OutputTokens

	_, err := m.db.ExecCtx(ctx, query,
		callLog.ModelId, callLog.ModelName, date, successCount, failedCount,
		callLog.InputTokens, callLog.OutputTokens, totalTokens, callLog.Cost, latencyMs,
		successCount, failedCount,
		callLog.InputTokens, callLog.OutputTokens, totalTokens, callLog.Cost, latencyMs,
	)

	return err
}

func (m *CostManager) checkCostAlerts(ctx context.Context, callLog *model.AmCallLog) error {
	query := "SELECT * FROM am_cost_alert_config WHERE status = 1 LIMIT 1"

	var config model.AmCostAlertConfig
	err := m.db.QueryRowCtx(ctx, &config, query)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil
		}
		return fmt.Errorf("query alert config: %w", err)
	}

	shouldAlert, alertType, actualValue, err := m.shouldTriggerAlert(ctx, &config, callLog)
	if err != nil {
		logx.Errorf("failed to check alert condition: %v", err)
		return nil
	}

	if shouldAlert {
		if err := m.createAlertRecord(ctx, &config, callLog, alertType, actualValue); err != nil {
			logx.Errorf("failed to create alert record: %v", err)
		}
	}

	return nil
}

func (m *CostManager) shouldTriggerAlert(ctx context.Context, config *model.AmCostAlertConfig, callLog *model.AmCallLog) (bool, string, float64, error) {
	now := time.Now()

	// Check daily limit
	if config.DailyLimit > 0 {
		startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		query := "SELECT COALESCE(SUM(total_cost), 0) FROM am_usage_statistics WHERE stat_date >= ?"

		var totalCost float64
		err := m.db.QueryRowCtx(ctx, &totalCost, query, startTime.Format("2006-01-02"))
		if err != nil {
			return false, "", 0, err
		}

		threshold := config.DailyLimit * config.AlertThreshold
		if totalCost >= threshold {
			return true, "daily", totalCost, nil
		}
	}

	// Check monthly limit
	if config.MonthlyLimit > 0 {
		startTime := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		query := "SELECT COALESCE(SUM(total_cost), 0) FROM am_usage_statistics WHERE stat_date >= ?"

		var totalCost float64
		err := m.db.QueryRowCtx(ctx, &totalCost, query, startTime.Format("2006-01-02"))
		if err != nil {
			return false, "", 0, err
		}

		threshold := config.MonthlyLimit * config.AlertThreshold
		if totalCost >= threshold {
			return true, "monthly", totalCost, nil
		}
	}

	return false, "", 0, nil
}

func (m *CostManager) createAlertRecord(ctx context.Context, config *model.AmCostAlertConfig, callLog *model.AmCallLog, alertType string, actualValue float64) error {
	lastAlertQuery := "SELECT alert_time FROM am_alert_record WHERE alert_type = ? AND status = 'pending' ORDER BY alert_time DESC LIMIT 1"

	var lastAlertTime time.Time
	err := m.db.QueryRowCtx(ctx, &lastAlertTime, lastAlertQuery, alertType)
	if err == nil {
		if time.Since(lastAlertTime) < 1*time.Hour {
			return nil
		}
	}

	var threshold float64
	if alertType == "daily" {
		threshold = config.DailyLimit * config.AlertThreshold
	} else {
		threshold = config.MonthlyLimit * config.AlertThreshold
	}

	message := fmt.Sprintf("Cost alert: %s cost $%.2f has reached threshold $%.2f (%.0f%%)",
		alertType, actualValue, threshold, config.AlertThreshold*100)

	query := `INSERT INTO am_alert_record
		(alert_type, threshold, actual_value, message, model_id, model_name, status)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = m.db.ExecCtx(ctx, query, alertType, threshold, actualValue, message,
		callLog.ModelId, callLog.ModelName, "pending")

	return err
}

func (m *CostManager) GetStatistics(ctx context.Context, modelID int64, startDate, endDate string) ([]*model.AmUsageStatistics, error) {
	query := "SELECT * FROM am_usage_statistics WHERE model_id = ? AND stat_date >= ? AND stat_date <= ? ORDER BY stat_date ASC"

	var stats []*model.AmUsageStatistics
	err := m.db.QueryRowsCtx(ctx, &stats, query, modelID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("query statistics: %w", err)
	}

	return stats, nil
}

func (m *CostManager) GetTotalCost(ctx context.Context, modelID int64, startDate, endDate string) (float64, error) {
	query := "SELECT COALESCE(SUM(total_cost), 0) FROM am_usage_statistics WHERE model_id = ? AND stat_date >= ? AND stat_date <= ?"

	var totalCost float64
	err := m.db.QueryRowCtx(ctx, &totalCost, query, modelID, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("query total cost: %w", err)
	}

	return totalCost, nil
}

func (m *CostManager) GetCallLogs(ctx context.Context, modelID int64, limit, offset int64) ([]*model.AmCallLog, error) {
	query := "SELECT * FROM am_call_log WHERE model_id = ? ORDER BY created_time DESC LIMIT ? OFFSET ?"

	var logs []*model.AmCallLog
	err := m.db.QueryRowsCtx(ctx, &logs, query, modelID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query call logs: %w", err)
	}

	return logs, nil
}
