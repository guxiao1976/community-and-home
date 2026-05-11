package auditlog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AuditLogger struct {
	db sqlx.SqlConn
}

func NewAuditLogger(db sqlx.SqlConn) *AuditLogger {
	return &AuditLogger{db: db}
}

type LogEntry struct {
	ContentType    string
	ContentSummary string
	RiskLevel      string
	Pass           bool
	Reason         string
	CheckLayer     string
	MatchedItems   []MatchedItem
	UserID         *int64
	SourceType     string
	SourceID       *int64
	NeedReview     bool
}

type MatchedItem struct {
	Layer       string  `json:"layer"`
	MatchedText string  `json:"matched_text"`
	Category    string  `json:"category"`
	Severity    int     `json:"severity"`
	Confidence  float64 `json:"confidence"`
}

func (l *AuditLogger) Log(ctx context.Context, entry LogEntry) error {
	matchedJSON, err := json.Marshal(entry.MatchedItems)
	if err != nil {
		return err
	}

	query := `INSERT INTO mod_audit_log
		(content_type, content_summary, risk_level, pass, reason, check_layer, matched_items,
		 user_id, source_type, source_id, need_review, created_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = l.db.ExecCtx(ctx, query,
		entry.ContentType,
		entry.ContentSummary,
		entry.RiskLevel,
		entry.Pass,
		entry.Reason,
		entry.CheckLayer,
		string(matchedJSON),
		entry.UserID,
		entry.SourceType,
		entry.SourceID,
		entry.NeedReview,
		time.Now(),
	)

	return err
}
