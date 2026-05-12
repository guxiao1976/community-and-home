package model

import (
	"context"
	"database/sql"
	"time"
)

type (
	// custom methods below
)

// InsertAuditLog is a convenience wrapper for creating audit log entries
func InsertAuditLog(ctx context.Context, m ModAuditLogModel, log *ModAuditLog) error {
	_, err := m.Insert(ctx, log)
	return err
}

// NewTextAuditLog creates an audit log for text review
func NewTextAuditLog(contentSummary, riskLevel string, pass bool, reason, checkLayer string, matchedItems string, userId int64, sourceType string, sourceId int64, needReview bool) *ModAuditLog {
	return &ModAuditLog{
		ContentType:   "text",
		ContentSummary: toNullString(contentSummary),
		RiskLevel:     riskLevel,
		Pass:          boolToInt(pass),
		Reason:        toNullString(reason),
		CheckLayer:    toNullString(checkLayer),
		MatchedItems:  toNullString(matchedItems),
		UserId:        toNullInt64(userId),
		SourceType:    toNullString(sourceType),
		SourceId:      toNullInt64(sourceId),
		NeedReview:    boolToInt(needReview),
		ReviewStatus:  0,
		CreatedTime:   time.Now(),
	}
}

// NewImageAuditLog creates an audit log for image review
func NewImageAuditLog(contentSummary, riskLevel string, pass bool, reason, checkLayer string, matchedItems string, userId int64) *ModAuditLog {
	return &ModAuditLog{
		ContentType:   "image",
		ContentSummary: toNullString(contentSummary),
		RiskLevel:     riskLevel,
		Pass:          boolToInt(pass),
		Reason:        toNullString(reason),
		CheckLayer:    toNullString(checkLayer),
		MatchedItems:  toNullString(matchedItems),
		UserId:        toNullInt64(userId),
		NeedReview:    boolToInt(false),
		ReviewStatus:  0,
		CreatedTime:   time.Now(),
	}
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func toNullInt64(i int64) sql.NullInt64 {
	return sql.NullInt64{Int64: i, Valid: true}
}
