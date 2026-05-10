package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type SubmissionRecord struct {
	Id             int64          `db:"id"`
	EntityType     string         `db:"entity_type"`
	EntityId       int64          `db:"entity_id"`
	EntityName     sql.NullString `db:"entity_name"`
	EntityCode     sql.NullString `db:"entity_code"`
	SubmissionType int64          `db:"submission_type"`
	SubmitterId    int64          `db:"submitter_id"`
	SubmitTime     time.Time      `db:"submit_time"`
	ReviewerId     sql.NullInt64  `db:"reviewer_id"`
	ReviewTime     sql.NullTime   `db:"review_time"`
	ReviewResult   int64          `db:"review_result"`
	ReviewNotes    sql.NullString `db:"review_notes"`
	CreatedTime    time.Time      `db:"created_time"`
}

type SubmissionRecordModel interface {
	Insert(ctx context.Context, record *SubmissionRecord) (int64, error)
	UpdateResult(ctx context.Context, id int64, reviewerId int64, result int64, notes string) error
	UpdateResultByEntity(ctx context.Context, entityType string, entityId int64, reviewerId int64, result int64, notes string) error
	FindById(ctx context.Context, id int64) (*SubmissionRecord, error)
	FindBySubmitter(ctx context.Context, submitterId int64, entityType *string, reviewResult *int64, page, pageSize int64) ([]*SubmissionRecord, int64, error)
	FindByReviewer(ctx context.Context, reviewerId int64, entityType *string, reviewResult *int64, page, pageSize int64) ([]*SubmissionRecord, int64, error)
}

type defaultSubmissionRecordModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewSubmissionRecordModel(conn sqlx.SqlConn) SubmissionRecordModel {
	return &defaultSubmissionRecordModel{
		conn:  conn,
		table: "md_submission_record",
	}
}

func (m *defaultSubmissionRecordModel) Insert(ctx context.Context, record *SubmissionRecord) (int64, error) {
	query := fmt.Sprintf("insert into %s (entity_type, entity_id, entity_name, entity_code, submission_type, submitter_id, submit_time, review_result) values (?, ?, ?, ?, ?, ?, ?, 0)", m.table)
	result, err := m.conn.ExecCtx(ctx, query,
		record.EntityType, record.EntityId, record.EntityName, record.EntityCode,
		record.SubmissionType, record.SubmitterId, record.SubmitTime)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (m *defaultSubmissionRecordModel) UpdateResult(ctx context.Context, id int64, reviewerId int64, result int64, notes string) error {
	query := fmt.Sprintf("update %s set reviewer_id = ?, review_time = now(), review_result = ?, review_notes = ? where id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, reviewerId, result, notes, id)
	return err
}

func (m *defaultSubmissionRecordModel) UpdateResultByEntity(ctx context.Context, entityType string, entityId int64, reviewerId int64, result int64, notes string) error {
	query := fmt.Sprintf("update %s set reviewer_id = ?, review_time = now(), review_result = ?, review_notes = ? where entity_type = ? and entity_id = ? and review_result = 0 order by id desc limit 1", m.table)
	_, err := m.conn.ExecCtx(ctx, query, reviewerId, result, notes, entityType, entityId)
	return err
}

func (m *defaultSubmissionRecordModel) FindById(ctx context.Context, id int64) (*SubmissionRecord, error) {
	var record SubmissionRecord
	query := fmt.Sprintf("select id, entity_type, entity_id, entity_name, entity_code, submission_type, submitter_id, submit_time, reviewer_id, review_time, review_result, review_notes, created_time from %s where id = ?", m.table)
	err := m.conn.QueryRowCtx(ctx, &record, query, id)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (m *defaultSubmissionRecordModel) FindBySubmitter(ctx context.Context, submitterId int64, entityType *string, reviewResult *int64, page, pageSize int64) ([]*SubmissionRecord, int64, error) {
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "submitter_id = ?")
	args = append(args, submitterId)
	if entityType != nil {
		conditions = append(conditions, "entity_type = ?")
		args = append(args, *entityType)
	}
	if reviewResult != nil {
		conditions = append(conditions, "review_result = ?")
		args = append(args, *reviewResult)
	}
	where := strings.Join(conditions, " and ")

	var count int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	err := m.conn.QueryRowCtx(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf("select id, entity_type, entity_id, entity_name, entity_code, submission_type, submitter_id, submit_time, reviewer_id, review_time, review_result, review_notes, created_time from %s where %s order by id desc limit ? offset ?", m.table, where)
	args = append(args, pageSize, offset)

	var records []*SubmissionRecord
	err = m.conn.QueryRowsCtx(ctx, &records, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	return records, count, nil
}

func (m *defaultSubmissionRecordModel) FindByReviewer(ctx context.Context, reviewerId int64, entityType *string, reviewResult *int64, page, pageSize int64) ([]*SubmissionRecord, int64, error) {
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "reviewer_id = ?")
	args = append(args, reviewerId)
	conditions = append(conditions, "review_result in (1, 2)")
	if entityType != nil {
		conditions = append(conditions, "entity_type = ?")
		args = append(args, *entityType)
	}
	if reviewResult != nil {
		conditions = append(conditions, "review_result = ?")
		args = append(args, *reviewResult)
	}
	where := strings.Join(conditions, " and ")

	var count int64
	countQuery := fmt.Sprintf("select count(*) from %s where %s", m.table, where)
	err := m.conn.QueryRowCtx(ctx, &count, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf("select id, entity_type, entity_id, entity_name, entity_code, submission_type, submitter_id, submit_time, reviewer_id, review_time, review_result, review_notes, created_time from %s where %s order by id desc limit ? offset ?", m.table, where)
	args = append(args, pageSize, offset)

	var records []*SubmissionRecord
	err = m.conn.QueryRowsCtx(ctx, &records, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	return records, count, nil
}
