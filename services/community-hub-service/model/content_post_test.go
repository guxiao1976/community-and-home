package model

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"database/sql/driver"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// =============================================================================
// Task 1.5 读查询：IsReviewComplete + FindListByCommunity + FindOneReviewComplete + FindMarquee
// =============================================================================

func TestIsReviewComplete(t *testing.T) {
	tests := []struct {
		name                string
		status              int64
		approvedAttachments int64
		attachmentCount     int64
		want                bool
	}{
		{"approved + 已审附件数==计数 → 完整", StatusApproved, 2, 2, true},
		{"approved + 无附件 → 恒完整", StatusApproved, 0, 0, true},
		{"approved + 已审附件数<计数 → 不完整", StatusApproved, 1, 2, false},
		{"approved + 已审附件数>计数 → 不完整", StatusApproved, 3, 2, false},
		{"draft → 不完整", StatusDraft, 0, 0, false},
		{"withdrawn → 不完整", StatusWithdrawn, 0, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsReviewComplete(tc.status, tc.approvedAttachments, tc.attachmentCount))
		})
	}
}

// contentPostCols 对齐 content_posts 全列（投影 content_posts.*, content_post_scope.community_id）。
func contentPostCols() []string {
	return []string{"id", "community_id", "title", "text", "role", "publisher", "publisher_id",
		"is_pinned", "published_at", "section_code", "status", "attachment_count",
		"kafka_push_status", "kafka_push_retries", "kafka_push_last_error", "kafka_pushed_at",
		"moderation_status", "moderation_time", "created_at", "updated_at", "deleted_at", "community_id"}
}

func contentPostRow(id, communityID int64, status int64, attCount int64) []driver.Value {
	// scope.community_id 列（末尾）取请求小区，覆盖 content_posts.community_id（弃用 NULL）
	return []driver.Value{id, nil, "标题", "正文", "community", "publisher", id, int32(0), time.Now(),
		"notice", status, attCount, int64(0), int64(0), nil, nil, int64(0), nil, time.Now(), time.Now(), nil, communityID}
}

func TestContentPostModel_FindListByCommunity(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	// count 查询
	mock.ExpectQuery("select count\\(\\*\\) from `content_posts` join `content_post_scope` on content_posts.id = content_post_scope.post_id.*").
		WithArgs(int64(2001), int64(StatusApproved)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	// 数据查询
	mock.ExpectQuery("select content_posts\\.\\*, content_post_scope\\.community_id from `content_posts` join `content_post_scope`.*order by content_posts\\.is_pinned desc, content_posts\\.published_at desc limit \\?, \\?").
		WithArgs(int64(2001), int64(StatusApproved), int64(0), int64(10)).
		WillReturnRows(sqlmock.NewRows(contentPostCols()).AddRow(contentPostRow(1001, 2001, StatusApproved, 0)...))

	list, total, err := m.FindListByCommunity(context.Background(), 2001, "", "", 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, list, 1)
	// 投影：scope.community_id 覆盖弃用 NULL 列 → CommunityId = 请求小区
	require.NotNil(t, list[0].CommunityId)
	assert.Equal(t, int64(2001), *list[0].CommunityId)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostModel_FindListByCommunity_Filter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	// 带 section_code + role 筛选
	mock.ExpectQuery("select count\\(\\*\\) from `content_posts`.*section_code = \\?.*role = \\?.*").
		WithArgs(int64(2001), int64(StatusApproved), "notice", "community").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("select content_posts\\.\\*.*section_code = \\?.*role = \\?.*limit \\?, \\?").
		WithArgs(int64(2001), int64(StatusApproved), "notice", "community", int64(10), int64(5)).
		WillReturnRows(sqlmock.NewRows(contentPostCols()))

	list, total, err := m.FindListByCommunity(context.Background(), 2001, "notice", "community", 10, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, list, 0)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostModel_FindOneReviewComplete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	mock.ExpectQuery("select \\* from `content_posts` where id = \\? and deleted_at is null and status = \\?.*").
		WithArgs(int64(1001), int64(StatusApproved)).
		WillReturnRows(sqlmock.NewRows(contentPostCols()).AddRow(contentPostRow(1001, 2001, StatusApproved, 0)...))

	p, err := m.FindOneReviewComplete(context.Background(), 1001)
	require.NoError(t, err)
	assert.Equal(t, int64(1001), p.Id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostModel_FindOneReviewComplete_NotComplete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	// draft 帖经谓词过滤返回 ErrNoRows（读路径不 mutate status，评审 M4）
	mock.ExpectQuery("select \\* from `content_posts` where id = \\? and deleted_at is null and status = \\?.*").
		WithArgs(int64(1001), int64(StatusApproved)).
		WillReturnError(sql.ErrNoRows)

	_, err = m.FindOneReviewComplete(context.Background(), 1001)
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

// =============================================================================
// Task 1.3 时间窗口：ContentPostListOption.WithTimeWindow + FindListByCommunity 窗口谓词
// =============================================================================

// TestContentPostListOption_WithTimeWindow — 选项语义：缺省 nil（不过滤）、WithTimeWindow(since) 返回下界。
func TestContentPostListOption_WithTimeWindow(t *testing.T) {
	// 无选项 → nil（additive：缺省不过滤）
	assert.Nil(t, ContentPostListOptionSince())

	// WithTimeWindow(s) → 返回下界 s（含端点）
	s := time.Date(2026, 7, 17, 0, 0, 0, 0, time.Local)
	got := ContentPostListOptionSince(WithTimeWindow(s))
	require.NotNil(t, got)
	assert.Equal(t, s, *got)
}

// TestContentPostModel_FindListByCommunity_WithWindow — 窗口谓词 sqlmock 锁定（RED: 新签名/谓词未实现）。
// 场景：窗口内行返回（count+list SQL 含 published_at >= ? AND published_at <= ?）；
// NULL/未来行排除由 SQL 谓词保证（published_at NULL 恒不匹配下界、未来行被上界排除）；
// 无窗口选项时 SQL 不含窗口谓词（additive，PC 管理列表行为不变）。
func TestContentPostModel_FindListByCommunity_WithWindow(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	t.Run("窗口内行返回 + NULL/未来行被 SQL 谓词排除", func(t *testing.T) {
		since := time.Date(2026, 7, 17, 0, 0, 0, 0, time.Local)
		// count 查询：窗口谓词在 section/role 之后（args: communityId, status, since, now）
		mock.ExpectQuery("select count\\(\\*\\) from `content_posts` join `content_post_scope`.*published_at >= \\? and content_posts.published_at <= \\?.*").
			WithArgs(int64(2001), int64(StatusApproved), since, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		// 数据查询：窗口谓词 + order by is_pinned desc, published_at desc（args: + since, now, offset, limit）
		mock.ExpectQuery("select content_posts\\.\\*, content_post_scope\\.community_id.*published_at >= \\? and content_posts.published_at <= \\?.*order by content_posts\\.is_pinned desc, content_posts\\.published_at desc limit \\?, \\?").
			WithArgs(int64(2001), int64(StatusApproved), since, sqlmock.AnyArg(), int64(0), int64(10)).
			WillReturnRows(sqlmock.NewRows(contentPostCols()).AddRow(contentPostRow(1001, 2001, StatusApproved, 0)...))

		list, total, err := m.FindListByCommunity(context.Background(), 2001, "", "", 0, 10, WithTimeWindow(since))
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, list, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("无窗口选项 → SQL 不含窗口谓词（缺省不过滤）", func(t *testing.T) {
		// 不加窗口谓词时 count/list 均仅 communityId + status 两个参数；
		// 若实现误加窗口谓词，WithArgs 参数个数不匹配 → 本子测试 FAIL。
		mock.ExpectQuery("select count\\(\\*\\) from `content_posts` join `content_post_scope`.*").
			WithArgs(int64(2001), int64(StatusApproved)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectQuery("select content_posts\\.\\*, content_post_scope\\.community_id.*limit \\?, \\?").
			WithArgs(int64(2001), int64(StatusApproved), int64(0), int64(10)).
			WillReturnRows(sqlmock.NewRows(contentPostCols()))

		list, total, err := m.FindListByCommunity(context.Background(), 2001, "", "", 0, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, list, 0)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestContentPostModel_FindMarquee(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	since := time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local)
	mock.ExpectQuery("select content_posts\\.\\*, content_post_scope\\.community_id.*published_at >= \\?.*limit \\?").
		WithArgs(int64(2001), int64(StatusApproved), since, int64(10)).
		WillReturnRows(sqlmock.NewRows(contentPostCols()).AddRow(contentPostRow(1001, 2001, StatusApproved, 0)...))

	list, err := m.FindMarquee(context.Background(), 2001, since, 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

// =============================================================================
// Task 1.6 写路径：Insert / UpdateContent / UpdateIsPinned / UpdateStatusAndPublish / Withdraw / UpdateKafkaPushStatus / FindPendingPush
// =============================================================================

func TestContentPostModel_Insert_ExplicitStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	// Insert SQL 不含 community_id 列，显式写 section_code/status/attachment_count/kafka_push_status；
	// published_at 由 submit 路径显式写（draft 时 sql.NullTime{} → NULL，运维验证发现的真实 bug 回归，见 _qa.md）
	mock.ExpectExec("insert into `content_posts` \\(id, title, `text`, role, publisher, publisher_id, is_pinned, section_code, status, attachment_count, kafka_push_status, moderation_status, moderation_time, published_at\\) values \\(\\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?\\)").
		WithArgs(int64(1001), "标题", "正文", "community", "publisher", int64(100), int32(0), "notice", int64(StatusDraft), int64(0), int64(KafkaPushNone), int64(0), sql.NullTime{}, sql.NullTime{}).
		WillReturnResult(sqlmock.NewResult(1, 1))

	n := &ContentPost{
		Id: 1001, Title: "标题", Text: "正文", Role: "community", Publisher: "publisher",
		PublisherId: int64Ptr(100), IsPinned: 0, SectionCode: "notice",
		Status: StatusDraft, AttachmentCount: 0, KafkaPushStatus: KafkaPushNone,
	}
	id, err := m.Insert(context.Background(), n)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

// TestContentPostModel_Insert_SubmitWritesPublishedAt — submit 路径 published_at=NOW() 必须落库
// （回归：运维验证发现 Insert SQL 漏 published_at 列，跑马灯恒空；submit=approved 时须写入）。
func TestContentPostModel_Insert_SubmitWritesPublishedAt(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	ts := time.Date(2026, 8, 16, 12, 0, 0, 0, time.Local)
	mock.ExpectExec("insert into `content_posts` \\(id, title, `text`, role, publisher, publisher_id, is_pinned, section_code, status, attachment_count, kafka_push_status, moderation_status, moderation_time, published_at\\) values \\(\\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?, \\?\\)").
		WithArgs(int64(1002), "标题", "正文", "property", "publisher", int64(100), int32(0), "notice", int64(StatusApproved), int64(0), int64(KafkaPushPending), int64(0), sql.NullTime{}, sql.NullTime{Valid: true, Time: ts}).
		WillReturnResult(sqlmock.NewResult(1, 1))

	n := &ContentPost{
		Id: 1002, Title: "标题", Text: "正文", Role: "property", Publisher: "publisher",
		PublisherId: int64Ptr(100), IsPinned: 0, SectionCode: "notice",
		Status: StatusApproved, AttachmentCount: 0, KafkaPushStatus: KafkaPushPending,
		PublishedAt: sql.NullTime{Valid: true, Time: ts},
	}
	id, err := m.Insert(context.Background(), n)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostModel_UpdateContent_DoesNotTouchStatusOrPinned(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	// 仅三列 title/text/section_code，无 status/is_pinned
	mock.ExpectExec("update `content_posts` set title = \\?, `text` = \\?, section_code = \\? where id = \\? and deleted_at is null").
		WithArgs("新标题", "新正文", "notice", int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = m.UpdateContent(context.Background(), 1001, "新标题", "新正文", "notice")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostModel_UpdateIsPinned_IndependentColumn(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	// 仅写 is_pinned 列，不碰 title/text/section_code（V5 修复）
	mock.ExpectExec("update `content_posts` set is_pinned = \\? where id = \\? and deleted_at is null").
		WithArgs(int32(1), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = m.UpdateIsPinned(context.Background(), 1001, 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostModel_UpdateStatusAndPublish(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	now := time.Now()
	// submit：status=approved + published_at + kafka_push_status=1 原子（单语句）
	mock.ExpectExec("update `content_posts` set status = \\?, published_at = \\?, kafka_push_status = \\? where id = \\? and deleted_at is null").
		WithArgs(int64(StatusApproved), now, int64(KafkaPushPending), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = m.UpdateStatusAndPublish(context.Background(), 1001, StatusApproved, now)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostModel_Withdraw(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	// 软删 + status=withdrawn 单语句原子
	mock.ExpectExec("update `content_posts` set deleted_at = now\\(\\), status = \\? where id = \\? and deleted_at is null").
		WithArgs(int64(StatusWithdrawn), int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = m.Withdraw(context.Background(), 1001)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostModel_UpdateKafkaPushStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	// 失败回调：保留 pending + retries + last_error 落库
	mock.ExpectExec("update `content_posts` set kafka_push_status = \\?, kafka_push_retries = \\?, kafka_push_last_error = \\?, kafka_pushed_at = \\? where id = \\?").
		WithArgs(int64(KafkaPushPending), int64(2), sql.NullString{String: "broker unreachable", Valid: true}, sql.NullTime{}, int64(1001)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = m.UpdateKafkaPushStatus(context.Background(), 1001, KafkaPushPending, 2,
		sql.NullString{String: "broker unreachable", Valid: true}, sql.NullTime{})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestContentPostModel_FindPendingPush(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	conn := sqlx.NewSqlConnFromDB(db)
	m := NewContentPostModel(conn)

	mock.ExpectQuery("select \\* from `content_posts` where kafka_push_status = \\? and deleted_at is null limit \\?").
		WithArgs(int64(KafkaPushPending), int64(100)).
		WillReturnRows(sqlmock.NewRows(contentPostCols()).AddRow(contentPostRow(1001, 2001, StatusApproved, 0)...))

	list, err := m.FindPendingPush(context.Background(), 100)
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func int64Ptr(v int64) *int64 { return &v }
