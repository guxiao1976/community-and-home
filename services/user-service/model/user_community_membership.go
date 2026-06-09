package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// UserCommunityMembership 小区成员关系表
type UserCommunityMembership struct {
	Id          int64        `db:"id"`
	UserId      int64        `db:"user_id"`
	CommunityId int64        `db:"community_id"`
	BindStatus  int64        `db:"bind_status"`
	JoinTime    time.Time    `db:"join_time"`
	LeaveTime   sql.NullTime `db:"leave_time"`
	CreatedTime time.Time    `db:"created_time"`
	UpdatedTime time.Time    `db:"updated_time"`
	Building    int          `db:"building"` // 楼号
	Unit        int          `db:"unit"`     // 单元号
	Room        int          `db:"room"`     // 房号
}

type UserCommunityMembershipModel interface {
	Insert(ctx context.Context, data *UserCommunityMembership) (sql.Result, error)
	FindOne(ctx context.Context, id int64) (*UserCommunityMembership, error)
	FindByUserAndCommunity(ctx context.Context, userId, communityId int64) (*UserCommunityMembership, error)
	FindByUserId(ctx context.Context, userId int64) ([]*UserCommunityMembership, error)
	CountActiveByUserId(ctx context.Context, userId int64) (int64, error)
	UpdateBindStatus(ctx context.Context, id int64, bindStatus int64, leaveTime time.Time) error
	// FindByAddress 按地址查询活跃成员（用于唯一性校验）
	FindByAddress(ctx context.Context, communityId int64, building, unit, room int) (*UserCommunityMembership, error)
	// UpdateAddress 更新地址信息（重新激活时使用）
	UpdateAddress(ctx context.Context, id int64, building, unit, room int) error
	// CountDistinctCommunities 用户历史上加入过的不同小区总数（所有状态）
	CountDistinctCommunities(ctx context.Context, userId int64) (int64, error)
	// CountDistinctCommunitiesThisYear 用户今年首次加入的不同小区数
	CountDistinctCommunitiesThisYear(ctx context.Context, userId int64, yearStart time.Time) (int64, error)
}

type defaultUserCommunityMembershipModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewUserCommunityMembershipModel(conn sqlx.SqlConn) UserCommunityMembershipModel {
	return &defaultUserCommunityMembershipModel{
		conn:  conn,
		table: "user_community_membership",
	}
}

func (m *defaultUserCommunityMembershipModel) Insert(ctx context.Context, data *UserCommunityMembership) (sql.Result, error) {
	query := fmt.Sprintf(`INSERT INTO %s (id, user_id, community_id, bind_status, join_time, created_time, updated_time, building, unit, room) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, m.table)
	return m.conn.ExecCtx(ctx, query, data.Id, data.UserId, data.CommunityId, data.BindStatus, data.JoinTime, data.CreatedTime, data.UpdatedTime, data.Building, data.Unit, data.Room)
}

func (m *defaultUserCommunityMembershipModel) FindOne(ctx context.Context, id int64) (*UserCommunityMembership, error) {
	query := fmt.Sprintf(`SELECT id, user_id, community_id, bind_status, join_time, leave_time, created_time, updated_time, building, unit, room FROM %s WHERE id = ?`, m.table)
	var resp UserCommunityMembership
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserCommunityMembershipModel) FindByUserAndCommunity(ctx context.Context, userId, communityId int64) (*UserCommunityMembership, error) {
	query := fmt.Sprintf(`SELECT id, user_id, community_id, bind_status, join_time, leave_time, created_time, updated_time, building, unit, room FROM %s WHERE user_id = ? AND community_id = ?`, m.table)
	var resp UserCommunityMembership
	err := m.conn.QueryRowCtx(ctx, &resp, query, userId, communityId)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserCommunityMembershipModel) FindByUserId(ctx context.Context, userId int64) ([]*UserCommunityMembership, error) {
	query := fmt.Sprintf(`SELECT id, user_id, community_id, bind_status, join_time, leave_time, created_time, updated_time, building, unit, room FROM %s WHERE user_id = ? AND bind_status = ?`, m.table)
	var resp []*UserCommunityMembership
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, MembershipBindStatusActive)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultUserCommunityMembershipModel) CountActiveByUserId(ctx context.Context, userId int64) (int64, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE user_id = ? AND bind_status = ?`, m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, userId, MembershipBindStatusActive)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *defaultUserCommunityMembershipModel) FindByAddress(ctx context.Context, communityId int64, building, unit, room int) (*UserCommunityMembership, error) {
	query := fmt.Sprintf(`SELECT id, user_id, community_id, bind_status, join_time, leave_time, created_time, updated_time, building, unit, room FROM %s WHERE community_id = ? AND building = ? AND unit = ? AND room = ? AND bind_status = ?`, m.table)
	var resp UserCommunityMembership
	err := m.conn.QueryRowCtx(ctx, &resp, query, communityId, building, unit, room, MembershipBindStatusActive)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultUserCommunityMembershipModel) UpdateAddress(ctx context.Context, id int64, building, unit, room int) error {
	query := fmt.Sprintf(`UPDATE %s SET building=?, unit=?, room=?, updated_time=? WHERE id=?`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, building, unit, room, time.Now(), id)
	return err
}

func (m *defaultUserCommunityMembershipModel) UpdateBindStatus(ctx context.Context, id int64, bindStatus int64, leaveTime time.Time) error {
	query := fmt.Sprintf(`UPDATE %s SET bind_status=?, leave_time=?, updated_time=? WHERE id=?`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, bindStatus, leaveTime, time.Now(), id)
	return err
}

func (m *defaultUserCommunityMembershipModel) CountDistinctCommunities(ctx context.Context, userId int64) (int64, error) {
	query := fmt.Sprintf(`SELECT COUNT(DISTINCT community_id) FROM %s WHERE user_id = ?`, m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, userId)
	return count, err
}

func (m *defaultUserCommunityMembershipModel) CountDistinctCommunitiesThisYear(ctx context.Context, userId int64, yearStart time.Time) (int64, error) {
	query := fmt.Sprintf(`SELECT COUNT(DISTINCT community_id) FROM %s WHERE user_id = ? AND join_time >= ?`, m.table)
	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, userId, yearStart)
	return count, err
}
