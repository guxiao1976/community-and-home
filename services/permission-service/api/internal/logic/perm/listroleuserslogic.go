package perm

import (
	"context"
	"fmt"

	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-permission/api/internal/svc"
	"github.com/guxiao1976/community-permission/api/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ListRoleUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListRoleUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRoleUsersLogic {
	return &ListRoleUsersLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

type roleUserRow struct {
	UserId   int64  `db:"user_id"`
	Phone    string `db:"phone"`
	Nickname string `db:"nickname"`
}

func (l *ListRoleUsersLogic) ListRoleUsers(req *types.ListRoleUsersReq) (*types.ListRoleUsersResp, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Count
	var total int64
	countSQL := `SELECT COUNT(*)
		FROM user.user_base u
		INNER JOIN permission.rel_user_role r ON r.user_id = u.id
		WHERE r.role_id = ? AND u.deleted_at IS NULL`
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &total, countSQL, req.Id); err != nil {
		return nil, err
	}

	// Query page
	var rows []roleUserRow
	err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows,
		`SELECT u.id as user_id, u.phone, u.nickname
		 FROM user.user_base u
		 INNER JOIN permission.rel_user_role r ON r.user_id = u.id
		 WHERE r.role_id = ? AND u.deleted_at IS NULL
		 ORDER BY u.id LIMIT ? OFFSET ?`, req.Id, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int32(0)
	if pageSize > 0 {
		totalPages = int32((total + int64(pageSize) - 1) / int64(pageSize))
	}

	users := make([]types.RoleUserInfo, 0, len(rows))
	for _, r := range rows {
		phone := r.Phone
		if decrypted, err := crypto.AESDecrypt(phone); err == nil {
			phone = decrypted
		}
		users = append(users, types.RoleUserInfo{
			UserId:   r.UserId,
			Phone:    phone,
			Nickname: r.Nickname,
		})
	}
	return &types.ListRoleUsersResp{
		Users:      users,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func init() { _ = fmt.Sprintf("") }
