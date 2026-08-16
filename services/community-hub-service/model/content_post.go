package model

import (
	"context"
	"database/sql"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 全生命周期 + 审核结果状态常量（REVISION 权威契约，对齐 Migration 003 status 列枚举）
const (
	StatusDraft     int64 = 0 // draft：草稿（可编辑）
	StatusSubmitted int64 = 1 // submitted：已提交（本期 submit 即隐式通过）
	StatusApproved  int64 = 2 // approved：审核通过（本期 submit 即置 2 + published_at=NOW()）
	StatusRejected  int64 = 3 // rejected：审核拒绝（仅审核流写入）
	StatusWithdrawn int64 = 4 // withdrawn：撤回（软删 + 置此状态）
)

// Kafka at-least-once 待推标记状态（对齐 Migration 003 kafka_push_status 列）
const (
	KafkaPushNone    int64 = 0 // 无待推
	KafkaPushPending int64 = 1 // pending-push（待重推）
	KafkaPushDone    int64 = 2 // 已推(ack)
)

// ContentPost 通用图文发布（原 notices RENAME，Migration 003）。
// community_id / published_at 为弃用/可空列：新写路径不写入，读路径经 scope JOIN 派生。
//
// SEE: [[restore-compensation-zero-time]] — PublishedAt 用 sql.NullTime，严禁 time.Time{} 零值写 DATETIME
type ContentPost struct {
	Id                 int64          `db:"id"`
	CommunityId        *int64         `db:"community_id"` // 弃用：范围关联单源 content_post_scope（兼容期保留列，不写入）
	Title              string         `db:"title"`
	Text               string         `db:"text"`         // 正文（原 content 改名；REST wire 仍以 content 键输出，见 design §REST wire 兼容）
	Role               string         `db:"role"`         // 发布角色（RBAC→映射派生：community/committee/property/grid_officer）
	Publisher          string         `db:"publisher"`    // 展示名（取用户真实档案，禁请求体信任）
	PublisherId        *int64         `db:"publisher_id"` // 发布人ID（JWT 派生）
	IsPinned           int32          `db:"is_pinned"`
	PublishedAt        sql.NullTime   `db:"published_at"`          // 审核锚定：本期 submit 即置 NOW()（去 NOT NULL）
	SectionCode        string         `db:"section_code"`          // 板块：notice=通知/repair=维修保修/...
	Status             int64          `db:"status"`                // 全生命周期+审核结果（StatusDraft..StatusWithdrawn）
	AttachmentCount    int64          `db:"attachment_count"`      // 附件计数（审核完整性判定载体）
	KafkaPushStatus    int64          `db:"kafka_push_status"`     // 0=无待推 1=pending-push 2=已推(ack)
	KafkaPushRetries   int64          `db:"kafka_push_retries"`    // 重推次数
	KafkaPushLastError sql.NullString `db:"kafka_push_last_error"` // 最近一次推送错误摘要（可观测）
	KafkaPushedAt      sql.NullTime   `db:"kafka_pushed_at"`       // 成功推送时间
	ModerationStatus   int64          `db:"moderation_status"`     // 兼容期保留（逐步过渡到 status + 附件级）
	ModerationTime     sql.NullTime   `db:"moderation_time"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
	DeletedAt          *time.Time     `db:"deleted_at"`
}

// ContentPostModel 通用图文发布数据访问层
type ContentPostModel interface {
	// Insert 显式写 section_code/status/attachment_count/kafka_push_status；
	// community_id/published_at 不写入（published_at 由 submit 路径 UpdateStatusAndPublish 显式传参）。
	Insert(ctx context.Context, n *ContentPost) (int64, error)
	// InsertTx 事务内插入（Create 单事务落库经共享 session）。
	InsertTx(ctx context.Context, session sqlx.Session, n *ContentPost) (int64, error)
	// FindOne 写接口存在性/归属校验（deleted_at IS NULL）。
	FindOne(ctx context.Context, id int64) (*ContentPost, error)
	// FindListByCommunity 读列表：JOIN content_post_scope（scope.community_id=?）+ IsReviewComplete 谓词 +
	// 可选 section_code/role 筛选 + order by is_pinned desc, published_at desc（NULLS LAST 防御）+ 分页。
	// 显式投影 content_posts.* + content_post_scope.community_id（右表限定，防双 community_id 列取到弃用 NULL）。
	FindListByCommunity(ctx context.Context, communityId int64, sectionCode, role string, offset, limit int64) ([]*ContentPost, int64, error)
	// FindOneReviewComplete 读详情专用：仅返回审核完整（status=approved + 附件完整性谓词）的内容。
	FindOneReviewComplete(ctx context.Context, id int64) (*ContentPost, error)
	// FindMarquee 跑马灯：JOIN scope、status=approved + 附件完整性、published_at >= since（含端点）、
	// order by is_pinned desc, published_at desc、limit。
	FindMarquee(ctx context.Context, communityId int64, since time.Time, limit int64) ([]*ContentPost, error)
	// UpdateContent draft 正文编辑：仅写 title/text/section_code 三列，不碰 status/is_pinned。
	// is_pinned 一律走 UpdateIsPinned（防「仅改附件/scope 传空 title/text」覆盖正文，评审 data-model v4 M1(c)）。
	UpdateContent(ctx context.Context, id int64, title, text, sectionCode string) error
	// UpdateContentTx 事务内 draft 正文编辑（Update 单事务经共享 session）。
	UpdateContentTx(ctx context.Context, session sqlx.Session, id int64, title, text, sectionCode string) error
	// UpdateIsPinned 独立列更新：仅写 is_pinned 列，不碰 title/text/section_code
	// （V5 修复——is_pinned-only 路径 draft/submitted/approved 置顶/取消置顶一律走本方法，
	//  禁止复用 UpdateContent 传空 title/text 把已发布帖正文清空）。
	UpdateIsPinned(ctx context.Context, id int64, isPinned int32) error
	// UpdateIsPinnedTx 事务内独立列更新（Update 单事务经共享 session）。
	UpdateIsPinnedTx(ctx context.Context, session sqlx.Session, id int64, isPinned int32) error
	// UpdateStatusAndPublish submit 动作：status=approved + published_at + kafka_push_status=1 原子（单语句）。
	UpdateStatusAndPublish(ctx context.Context, id int64, status int64, publishedAt time.Time) error
	// UpdateStatusAndPublishTx 事务内 submit 动作（Update 单事务经共享 session）。
	UpdateStatusAndPublishTx(ctx context.Context, session sqlx.Session, id int64, status int64, publishedAt time.Time) error
	// UpdateAttachmentCount 同事务重算附件计数（draft 编辑附件集合全量替换后；空集=0，D19 不变量可归零）。
	UpdateAttachmentCount(ctx context.Context, id int64, count int64) error
	// UpdateAttachmentCountTx 事务内重算附件计数（Update 单事务经共享 session）。
	UpdateAttachmentCountTx(ctx context.Context, session sqlx.Session, id int64, count int64) error
	// Withdraw 撤回：软删 + status=withdrawn 单语句原子（整体回滚无半态；scope/附件行保留由调用方表达）。
	Withdraw(ctx context.Context, id int64) error
	// UpdateKafkaPushStatus 重推扫描器回调（ack 置 2 / 失败保留 pending + last_error 落库）。
	UpdateKafkaPushStatus(ctx context.Context, id int64, pushStatus, retries int64, lastErr sql.NullString, pushedAt sql.NullTime) error
	// FindPendingPush 重推扫描：kafka_push_status=1 且 deleted_at IS NULL。
	FindPendingPush(ctx context.Context, limit int64) ([]*ContentPost, error)
}

type defaultContentPostModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewContentPostModel(conn sqlx.SqlConn) ContentPostModel {
	return &defaultContentPostModel{conn: conn, table: "`content_posts`"}
}

// IsReviewComplete 审核完整性单一谓词（REQ-CPB-8 读路径单源）：
// 正文 status==approved 且 已审附件数（review_status=approved）== attachment_count。
// 列表 / 详情 / 跑马灯读查询共用；读路径绝不 mutate status。
func IsReviewComplete(status int64, approvedAttachments, attachmentCount int64) bool {
	return status == StatusApproved && approvedAttachments == attachmentCount
}

// attachmentCompleteSubquery 附件完整性关联标量子查询（count(attachments WHERE review_status=approved)==attachment_count）。
// 走 `idx_notice(post_id)`，勿退化为全表扫描（评审 data-model v4 I4）。
const attachmentCompleteSubquery = `(select count(*) from ` + "`content_post_attachments`" + ` where post_id = content_posts.id and review_status = 1) = content_posts.attachment_count`

func (m *defaultContentPostModel) Insert(ctx context.Context, n *ContentPost) (int64, error) {
	return m.insert(ctx, m.conn, n)
}

func (m *defaultContentPostModel) InsertTx(ctx context.Context, session sqlx.Session, n *ContentPost) (int64, error) {
	return m.insert(ctx, session, n)
}

func (m *defaultContentPostModel) insert(ctx context.Context, execer interface {
	ExecCtx(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}, n *ContentPost) (int64, error) {
	query := `insert into ` + m.table + `
		(id, title, ` + "`text`" + `, role, publisher, publisher_id, is_pinned, section_code, status, attachment_count, kafka_push_status, moderation_status, moderation_time)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := execer.ExecCtx(ctx, query,
		n.Id, n.Title, n.Text, n.Role,
		n.Publisher, n.PublisherId, n.IsPinned, n.SectionCode, n.Status, n.AttachmentCount,
		n.KafkaPushStatus, n.ModerationStatus, n.ModerationTime)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *defaultContentPostModel) FindOne(ctx context.Context, id int64) (*ContentPost, error) {
	var v ContentPost
	query := `select * from ` + m.table + ` where id = ? and deleted_at is null`
	err := m.conn.QueryRowCtx(ctx, &v, query, id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// FindListByCommunity 读列表：JOIN content_post_scope 按目标小区过滤 + 审核完整性谓词。
// 显式投影 `content_posts.*, content_post_scope.community_id`：go-zero sqlx 按列名赋值、
// 同名列后者覆盖前者——右表 scope.community_id 在投影末尾，确保 ContentPost.CommunityId 取到
// 请求小区而非 content_posts 弃用 NULL 列（评审 data-model v4 I4/JOIN 投影）。
func (m *defaultContentPostModel) FindListByCommunity(ctx context.Context, communityId int64, sectionCode, role string, offset, limit int64) ([]*ContentPost, int64, error) {
	var total int64
	countQuery := `select count(*) from ` + m.table + ` join ` + "`content_post_scope`" + ` on content_posts.id = content_post_scope.post_id
		where content_post_scope.community_id = ? and content_posts.deleted_at is null
		and content_posts.status = ? and ` + attachmentCompleteSubquery
	countArgs := []interface{}{communityId, StatusApproved}

	if sectionCode != "" {
		countQuery += ` and content_posts.section_code = ?`
		countArgs = append(countArgs, sectionCode)
	}
	if role != "" {
		countQuery += ` and content_posts.role = ?`
		countArgs = append(countArgs, role)
	}

	if err := m.conn.QueryRowCtx(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	query := `select content_posts.*, content_post_scope.community_id from ` + m.table + ` join ` + "`content_post_scope`" + ` on content_posts.id = content_post_scope.post_id
		where content_post_scope.community_id = ? and content_posts.deleted_at is null
		and content_posts.status = ? and ` + attachmentCompleteSubquery
	queryArgs := []interface{}{communityId, StatusApproved}

	if sectionCode != "" {
		query += ` and content_posts.section_code = ?`
		queryArgs = append(queryArgs, sectionCode)
	}
	if role != "" {
		query += ` and content_posts.role = ?`
		queryArgs = append(queryArgs, role)
	}
	query += ` order by content_posts.is_pinned desc, content_posts.published_at desc limit ?, ?`
	queryArgs = append(queryArgs, offset, limit)

	var list []*ContentPost
	err := m.conn.QueryRowsCtx(ctx, &list, query, queryArgs...)
	return list, total, err
}

// FindOneReviewComplete 读详情专用：仅返回审核完整（status=approved + 附件完整性谓词）的内容。
func (m *defaultContentPostModel) FindOneReviewComplete(ctx context.Context, id int64) (*ContentPost, error) {
	var v ContentPost
	query := `select * from ` + m.table + ` where id = ? and deleted_at is null
		and status = ? and ` + attachmentCompleteSubquery
	err := m.conn.QueryRowCtx(ctx, &v, query, id, StatusApproved)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// FindMarquee 跑马灯：JOIN scope、status=approved + 附件完整性、published_at >= since（含端点）。
func (m *defaultContentPostModel) FindMarquee(ctx context.Context, communityId int64, since time.Time, limit int64) ([]*ContentPost, error) {
	query := `select content_posts.*, content_post_scope.community_id from ` + m.table + ` join ` + "`content_post_scope`" + ` on content_posts.id = content_post_scope.post_id
		where content_post_scope.community_id = ? and content_posts.deleted_at is null
		and content_posts.status = ? and content_posts.published_at >= ? and ` + attachmentCompleteSubquery + ` order by content_posts.is_pinned desc, content_posts.published_at desc limit ?`
	var list []*ContentPost
	err := m.conn.QueryRowsCtx(ctx, &list, query, communityId, StatusApproved, since, limit)
	return list, err
}

// UpdateContent draft 正文编辑：仅写 title/text/section_code，不碰 status/is_pinned。
func (m *defaultContentPostModel) UpdateContent(ctx context.Context, id int64, title, text, sectionCode string) error {
	return m.updateContent(ctx, m.conn, id, title, text, sectionCode)
}

func (m *defaultContentPostModel) UpdateContentTx(ctx context.Context, session sqlx.Session, id int64, title, text, sectionCode string) error {
	return m.updateContent(ctx, session, id, title, text, sectionCode)
}

func (m *defaultContentPostModel) updateContent(ctx context.Context, e execer, id int64, title, text, sectionCode string) error {
	query := `update ` + m.table + ` set title = ?, ` + "`text`" + ` = ?, section_code = ? where id = ? and deleted_at is null`
	_, err := e.ExecCtx(ctx, query, title, text, sectionCode, id)
	return err
}

// UpdateIsPinned 独立列更新：仅写 is_pinned 列，不碰 title/text/section_code。
// is_pinned-only 路径（draft/submitted/approved 置顶/取消置顶）一律走本方法。
func (m *defaultContentPostModel) UpdateIsPinned(ctx context.Context, id int64, isPinned int32) error {
	return m.updateIsPinned(ctx, m.conn, id, isPinned)
}

func (m *defaultContentPostModel) UpdateIsPinnedTx(ctx context.Context, session sqlx.Session, id int64, isPinned int32) error {
	return m.updateIsPinned(ctx, session, id, isPinned)
}

func (m *defaultContentPostModel) updateIsPinned(ctx context.Context, e execer, id int64, isPinned int32) error {
	query := `update ` + m.table + ` set is_pinned = ? where id = ? and deleted_at is null`
	_, err := e.ExecCtx(ctx, query, isPinned, id)
	return err
}

// UpdateStatusAndPublish submit 动作：status + published_at + kafka_push_status=1 单语句原子（D16/D20）。
func (m *defaultContentPostModel) UpdateStatusAndPublish(ctx context.Context, id int64, status int64, publishedAt time.Time) error {
	return m.updateStatusAndPublish(ctx, m.conn, id, status, publishedAt)
}

func (m *defaultContentPostModel) UpdateStatusAndPublishTx(ctx context.Context, session sqlx.Session, id int64, status int64, publishedAt time.Time) error {
	return m.updateStatusAndPublish(ctx, session, id, status, publishedAt)
}

func (m *defaultContentPostModel) updateStatusAndPublish(ctx context.Context, e execer, id int64, status int64, publishedAt time.Time) error {
	query := `update ` + m.table + ` set status = ?, published_at = ?, kafka_push_status = ? where id = ? and deleted_at is null`
	_, err := e.ExecCtx(ctx, query, status, publishedAt, KafkaPushPending, id)
	return err
}

// UpdateAttachmentCount 重算附件计数（draft 编辑附件集合全量替换后）。
func (m *defaultContentPostModel) UpdateAttachmentCount(ctx context.Context, id int64, count int64) error {
	return m.updateAttachmentCount(ctx, m.conn, id, count)
}

func (m *defaultContentPostModel) UpdateAttachmentCountTx(ctx context.Context, session sqlx.Session, id int64, count int64) error {
	return m.updateAttachmentCount(ctx, session, id, count)
}

func (m *defaultContentPostModel) updateAttachmentCount(ctx context.Context, e execer, id int64, count int64) error {
	query := `update ` + m.table + ` set attachment_count = ? where id = ? and deleted_at is null`
	_, err := e.ExecCtx(ctx, query, count, id)
	return err
}

// Withdraw 撤回：软删 + status=withdrawn 单语句原子（无半态；scope/附件行保留由调用方表达，REQ-CPB-10）。
func (m *defaultContentPostModel) Withdraw(ctx context.Context, id int64) error {
	query := `update ` + m.table + ` set deleted_at = now(), status = ? where id = ? and deleted_at is null`
	_, err := m.conn.ExecCtx(ctx, query, StatusWithdrawn, id)
	return err
}

// UpdateKafkaPushStatus 重推扫描器回调：成功置 2 + pushed_at；失败保留 pending + retries/last_error 落库。
func (m *defaultContentPostModel) UpdateKafkaPushStatus(ctx context.Context, id int64, pushStatus, retries int64, lastErr sql.NullString, pushedAt sql.NullTime) error {
	query := `update ` + m.table + ` set kafka_push_status = ?, kafka_push_retries = ?, kafka_push_last_error = ?, kafka_pushed_at = ? where id = ?`
	_, err := m.conn.ExecCtx(ctx, query, pushStatus, retries, lastErr, pushedAt, id)
	return err
}

// FindPendingPush 重推扫描：kafka_push_status=1 且 deleted_at IS NULL。
func (m *defaultContentPostModel) FindPendingPush(ctx context.Context, limit int64) ([]*ContentPost, error) {
	var list []*ContentPost
	query := `select * from ` + m.table + ` where kafka_push_status = ? and deleted_at is null limit ?`
	err := m.conn.QueryRowsCtx(ctx, &list, query, KafkaPushPending, limit)
	return list, err
}
