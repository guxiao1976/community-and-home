package model

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// ==================== AuthCredential ====================

// AuthCredential 登录凭证表（spec/auth.md 数据模型）
type AuthCredential struct {
	Id           int64     `db:"id"`
	UserId       int64     `db:"user_id"`
	IdentityType string    `db:"identity_type"` // phone / sms / wechat
	Identifier   string    `db:"identifier"`    // RSA 密文存储的手机号
	Credential   string    `db:"credential"`    // bcrypt 密文
	CreatedTime  time.Time `db:"created_time"`
	UpdatedTime  time.Time `db:"updated_time"`
}

type AuthCredentialModel interface {
	Insert(ctx context.Context, data *AuthCredential) (int64, error)
	FindOne(ctx context.Context, id int64) (*AuthCredential, error)
	FindByIdentityTypeAndIdentifier(ctx context.Context, identityType, identifier string) (*AuthCredential, error)
	FindByUserId(ctx context.Context, userId int64) ([]*AuthCredential, error)
	UpdateCredential(ctx context.Context, id int64, newCredential string) error
}

type defaultAuthCredentialModel struct {
	conn  sqlx.SqlConn
	table string
}

func NewAuthCredentialModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) AuthCredentialModel {
	return &defaultAuthCredentialModel{conn: conn, table: "`auth_credential`"}
}

// Insert 插入新凭证
func (m *defaultAuthCredentialModel) Insert(ctx context.Context, data *AuthCredential) (int64, error) {
	query := fmt.Sprintf("insert into %s (user_id, identity_type, identifier, credential) values (?, ?, ?, ?)", m.table)
	res, err := m.conn.ExecCtx(ctx, query, data.UserId, data.IdentityType, data.Identifier, data.Credential)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// FindOne 根据 ID 查询
func (m *defaultAuthCredentialModel) FindOne(ctx context.Context, id int64) (*AuthCredential, error) {
	var v AuthCredential
	err := m.conn.QueryRowCtx(ctx, &v, fmt.Sprintf("select * from %s where id = ?", m.table), id)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// FindByIdentityTypeAndIdentifier 根据身份类型+标识查询（登录时使用）
func (m *defaultAuthCredentialModel) FindByIdentityTypeAndIdentifier(ctx context.Context, identityType, identifier string) (*AuthCredential, error) {
	var v AuthCredential
	query := fmt.Sprintf("select * from %s where identity_type = ? and identifier = ? limit 1", m.table)
	err := m.conn.QueryRowCtx(ctx, &v, query, identityType, identifier)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// FindByUserId 查询用户的所有凭证
func (m *defaultAuthCredentialModel) FindByUserId(ctx context.Context, userId int64) ([]*AuthCredential, error) {
	var list []*AuthCredential
	query := fmt.Sprintf("select * from %s where user_id = ?", m.table)
	err := m.conn.QueryRowsCtx(ctx, &list, query, userId)
	return list, err
}

// UpdateCredential 更新密码（bcrypt 密文）
func (m *defaultAuthCredentialModel) UpdateCredential(ctx context.Context, id int64, newCredential string) error {
	query := fmt.Sprintf("update %s set credential = ?, updated_time = now() where id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, newCredential, id)
	return err
}

var _ = time.Now
