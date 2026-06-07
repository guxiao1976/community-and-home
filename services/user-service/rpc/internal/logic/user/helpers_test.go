package user

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-user/model"
	"github.com/guxiao1976/community-user/rpc/internal/svc"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// sql.Result 模拟
// =============================================================================

type mockSQLResult struct{}

func (m mockSQLResult) LastInsertId() (int64, error) { return 1, nil }
func (m mockSQLResult) RowsAffected() (int64, error)  { return 1, nil }

var _ sql.Result = mockSQLResult{}

// =============================================================================
// Mock UserBaseModel
// =============================================================================

type mockUserBaseModel struct {
	data  map[int64]*model.UserBase
	phone map[string]*model.UserBase
}

func newMockUserBaseModel() *mockUserBaseModel {
	return &mockUserBaseModel{data: make(map[int64]*model.UserBase), phone: make(map[string]*model.UserBase)}
}

func (m *mockUserBaseModel) Insert(ctx context.Context, u *model.UserBase) (sql.Result, error) {
	m.data[u.Id] = u
	m.phone[u.Phone] = u
	return mockSQLResult{}, nil
}
func (m *mockUserBaseModel) FindOne(ctx context.Context, id int64) (*model.UserBase, error) {
	if u, ok := m.data[id]; ok {
		return u, nil
	}
	return nil, model.ErrNotFound
}
func (m *mockUserBaseModel) FindOneByPhone(ctx context.Context, phone string) (*model.UserBase, error) {
	if u, ok := m.phone[phone]; ok {
		return u, nil
	}
	return nil, model.ErrNotFound
}
func (m *mockUserBaseModel) FindByIds(ctx context.Context, ids []int64) ([]*model.UserBase, error) {
	var r []*model.UserBase
	for _, id := range ids {
		if u, ok := m.data[id]; ok {
			r = append(r, u)
		}
	}
	return r, nil
}
func (m *mockUserBaseModel) FindPage(ctx context.Context, kw string, status *int64, page, size int32) ([]*model.UserBase, int64, error) {
	var r []*model.UserBase
	for _, u := range m.data {
		r = append(r, u)
	}
	return r, int64(len(r)), nil
}
func (m *mockUserBaseModel) Update(ctx context.Context, u *model.UserBase) error { m.data[u.Id] = u; return nil }
func (m *mockUserBaseModel) SoftDelete(ctx context.Context, id int64) error {
	if u, ok := m.data[id]; ok {
		u.DeleteTime = sql.NullTime{Time: time.Now(), Valid: true}
	}
	return nil
}
func (m *mockUserBaseModel) UpdateStatus(ctx context.Context, id int64, s int64) error {
	if u, ok := m.data[id]; ok {
		u.Status = s
	}
	return nil
}
func (m *mockUserBaseModel) UpdateRealNameAndIdCard(ctx context.Context, id int64, realName, idCard string) error {
	u, ok := m.data[id]
	if !ok {
		return fmt.Errorf("user not found")
	}
	// COALESCE: first write only
	if u.RealName.String == "" || !u.RealName.Valid {
		u.RealName = sql.NullString{String: realName, Valid: true}
	}
	if u.IdCardNumber.String == "" || !u.IdCardNumber.Valid {
		u.IdCardNumber = sql.NullString{String: idCard, Valid: true}
	}
	return nil
}

var _ model.UserBaseModel = (*mockUserBaseModel)(nil)

// =============================================================================
// Mock UserCommunityMembershipModel
// =============================================================================

type mockMembershipModel struct {
	data          map[int64]*model.UserCommunityMembership
	byUserCommIdx map[string]int64
}

func newMockMembershipModel() *mockMembershipModel {
	return &mockMembershipModel{data: make(map[int64]*model.UserCommunityMembership), byUserCommIdx: make(map[string]int64)}
}

func (m *mockMembershipModel) Insert(ctx context.Context, d *model.UserCommunityMembership) (sql.Result, error) {
	m.data[d.Id] = d
	m.byUserCommIdx[fmt.Sprintf("%d_%d", d.UserId, d.CommunityId)] = d.Id
	return mockSQLResult{}, nil
}
func (m *mockMembershipModel) FindOne(ctx context.Context, id int64) (*model.UserCommunityMembership, error) {
	if d, ok := m.data[id]; ok {
		return d, nil
	}
	return nil, model.ErrNotFound
}
func (m *mockMembershipModel) FindByUserAndCommunity(ctx context.Context, uid, cid int64) (*model.UserCommunityMembership, error) {
	if id, ok := m.byUserCommIdx[fmt.Sprintf("%d_%d", uid, cid)]; ok {
		return m.data[id], nil
	}
	return nil, model.ErrNotFound
}
func (m *mockMembershipModel) FindByUserId(ctx context.Context, uid int64) ([]*model.UserCommunityMembership, error) {
	var r []*model.UserCommunityMembership
	for _, d := range m.data {
		if d.UserId == uid && d.BindStatus == model.MembershipBindStatusActive {
			r = append(r, d)
		}
	}
	return r, nil
}
func (m *mockMembershipModel) CountActiveByUserId(ctx context.Context, uid int64) (int64, error) {
	var n int64
	for _, d := range m.data {
		if d.UserId == uid && d.BindStatus == model.MembershipBindStatusActive {
			n++
		}
	}
	return n, nil
}
func (m *mockMembershipModel) UpdateBindStatus(ctx context.Context, id int64, s int64, lt time.Time) error {
	if d, ok := m.data[id]; ok {
		d.BindStatus = s
		if lt.IsZero() {
			d.LeaveTime = sql.NullTime{}
		} else {
			d.LeaveTime = sql.NullTime{Time: lt, Valid: true}
		}
	}
	return nil
}

func (m *mockMembershipModel) FindByAddress(ctx context.Context, communityId int64, building, unit, room int) (*model.UserCommunityMembership, error) {
	for _, d := range m.data {
		if d.CommunityId == communityId && d.Building == building && d.Unit == unit && d.Room == room && d.BindStatus == model.MembershipBindStatusActive {
			return d, nil
		}
	}
	return nil, model.ErrNotFound
}
func (m *mockMembershipModel) UpdateAddress(ctx context.Context, id int64, building, unit, room int) error {
	if d, ok := m.data[id]; ok {
		d.Building = building
		d.Unit = unit
		d.Room = room
	}
	return nil
}

var _ model.UserCommunityMembershipModel = (*mockMembershipModel)(nil)

// =============================================================================
// Mock UserMembershipRoleModel
// =============================================================================

type mockRoleModel struct {
	data map[int64]*model.UserMembershipRole
}

func newMockRoleModel() *mockRoleModel {
	return &mockRoleModel{data: make(map[int64]*model.UserMembershipRole)}
}

func (m *mockRoleModel) Insert(ctx context.Context, d *model.UserMembershipRole) (sql.Result, error) {
	m.data[d.Id] = d
	return mockSQLResult{}, nil
}
func (m *mockRoleModel) FindOne(ctx context.Context, id int64) (*model.UserMembershipRole, error) {
	if d, ok := m.data[id]; ok {
		return d, nil
	}
	return nil, model.ErrNotFound
}
func (m *mockRoleModel) FindByMembershipAndRole(ctx context.Context, mid int64, roleCode string) (*model.UserMembershipRole, error) {
	for _, d := range m.data {
		if d.MembershipId.Valid && d.MembershipId.Int64 == mid && d.RoleCode == roleCode {
			return d, nil
		}
	}
	return nil, model.ErrNotFound
}
func (m *mockRoleModel) FindByUserAndCommunity(ctx context.Context, uid, cid int64) ([]*model.UserMembershipRole, error) {
	var r []*model.UserMembershipRole
	for _, d := range m.data {
		if d.UserId == uid && d.CommunityId == cid {
			r = append(r, d)
		}
	}
	return r, nil
}
func (m *mockRoleModel) FindByUserId(ctx context.Context, uid int64) ([]*model.UserMembershipRole, error) {
	var r []*model.UserMembershipRole
	for _, d := range m.data {
		if d.UserId == uid {
			r = append(r, d)
		}
	}
	return r, nil
}
func (m *mockRoleModel) FindApprovedByUser(ctx context.Context, uid, cid int64, roleCodes []string) ([]*model.UserMembershipRole, error) {
	var r []*model.UserMembershipRole
	for _, d := range m.data {
		if d.UserId == uid && d.VerfStatus == model.RoleVerfStatusApproved {
			if cid > 0 && d.CommunityId != cid {
				continue
			}
			if len(roleCodes) > 0 {
				f := false
				for _, rc := range roleCodes {
					if d.RoleCode == rc {
						f = true
						break
					}
				}
				if !f {
					continue
				}
			}
			r = append(r, d)
		}
	}
	return r, nil
}
func (m *mockRoleModel) FindExpiredRoles(ctx context.Context) ([]*model.UserMembershipRole, error) {
	var r []*model.UserMembershipRole
	now := time.Now()
	for _, d := range m.data {
		if d.VerfStatus == model.RoleVerfStatusApproved && d.ExpiresAt.Valid && d.ExpiresAt.Time.Before(now) {
			r = append(r, d)
		}
	}
	return r, nil
}
func (m *mockRoleModel) UpdateVerfStatus(ctx context.Context, id int64, s int64, va, ex sql.NullTime) error {
	if d, ok := m.data[id]; ok {
		d.VerfStatus = s
		d.VerifiedAt = va
		d.ExpiresAt = ex
	}
	return nil
}
func (m *mockRoleModel) UpdateVerfStatusOnly(ctx context.Context, id int64, s int64) error {
	if d, ok := m.data[id]; ok {
		d.VerfStatus = s
	}
	return nil
}

var _ model.UserMembershipRoleModel = (*mockRoleModel)(nil)

// =============================================================================
// Mock UserCertificationModel
// =============================================================================

type mockCertModel struct {
	data map[int64]*model.UserCertification
}

func newMockCertModel() *mockCertModel { return &mockCertModel{data: make(map[int64]*model.UserCertification)} }

func (m *mockCertModel) Insert(ctx context.Context, d *model.UserCertification) (sql.Result, error) {
	m.data[d.Id] = d
	return mockSQLResult{}, nil
}
func (m *mockCertModel) FindOne(ctx context.Context, id int64) (*model.UserCertification, error) {
	if d, ok := m.data[id]; ok {
		return d, nil
	}
	return nil, model.ErrNotFound
}
func (m *mockCertModel) FindByRoleId(ctx context.Context, rid int64) ([]*model.UserCertification, error) {
	var r []*model.UserCertification
	for _, d := range m.data {
		if d.RoleId == rid {
			r = append(r, d)
		}
	}
	return r, nil
}
func (m *mockCertModel) FindByUserId(ctx context.Context, uid int64) ([]*model.UserCertification, error) {
	var r []*model.UserCertification
	for _, d := range m.data {
		if d.UserId == uid {
			r = append(r, d)
		}
	}
	return r, nil
}
func (m *mockCertModel) FindPage(ctx context.Context, status, uid *int64, page, size int32) ([]*model.UserCertification, int64, error) {
	var r []*model.UserCertification
	for _, d := range m.data {
		r = append(r, d)
	}
	return r, int64(len(r)), nil
}
func (m *mockCertModel) Update(ctx context.Context, d *model.UserCertification) error {
	m.data[d.Id] = d
	return nil
}

var _ model.UserCertificationModel = (*mockCertModel)(nil)

// =============================================================================
// Mock UserResidenceModel
// =============================================================================

type mockResidenceModel struct {
	data map[int64]*model.UserResidence
}

func newMockResidenceModel() *mockResidenceModel {
	return &mockResidenceModel{data: make(map[int64]*model.UserResidence)}
}

func (m *mockResidenceModel) Insert(ctx context.Context, d *model.UserResidence) (sql.Result, error) {
	m.data[d.Id] = d
	return mockSQLResult{}, nil
}
func (m *mockResidenceModel) FindByMembershipId(ctx context.Context, mid int64) ([]*model.UserResidence, error) {
	var r []*model.UserResidence
	for _, d := range m.data {
		if d.MembershipId == mid {
			r = append(r, d)
		}
	}
	return r, nil
}
func (m *mockResidenceModel) FindByUserId(ctx context.Context, uid int64) ([]*model.UserResidence, error) {
	var r []*model.UserResidence
	for _, d := range m.data {
		if d.UserId == uid {
			r = append(r, d)
		}
	}
	return r, nil
}
func (m *mockResidenceModel) FindByMembershipAndHouse(ctx context.Context, mid int64, houseId string) (*model.UserResidence, error) {
	for _, d := range m.data {
		if d.MembershipId == mid && d.HouseId == houseId {
			return d, nil
		}
	}
	return nil, model.ErrNotFound
}
func (m *mockResidenceModel) Update(ctx context.Context, d *model.UserResidence) error {
	m.data[d.Id] = d
	return nil
}

var _ model.UserResidenceModel = (*mockResidenceModel)(nil)

// =============================================================================
// Test setup
// =============================================================================

const testAESKey = "1234567890abcdef1234567890abcdef"

// testSvc creates a ServiceContext with all mock models.
func testSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()
	require.NoError(t, crypto.InitAES(testAESKey))
	return &svc.ServiceContext{
		UserBaseModel:                 newMockUserBaseModel(),
		UserCommunityMembershipModel:  newMockMembershipModel(),
		UserMembershipRoleModel:       newMockRoleModel(),
		UserCertificationModel:        newMockCertModel(),
		UserResidenceModel:            newMockResidenceModel(),
		Redis:                         nil, // nil = cache bypass
	}
}

// cast helpers — return typed mock from svcCtx for direct data access
func userBaseModel(s *svc.ServiceContext) *mockUserBaseModel       { return s.UserBaseModel.(*mockUserBaseModel) }
func membershipModel(s *svc.ServiceContext) *mockMembershipModel   { return s.UserCommunityMembershipModel.(*mockMembershipModel) }
func roleModel(s *svc.ServiceContext) *mockRoleModel             { return s.UserMembershipRoleModel.(*mockRoleModel) }
func certModel(s *svc.ServiceContext) *mockCertModel             { return s.UserCertificationModel.(*mockCertModel) }
func residenceModel(s *svc.ServiceContext) *mockResidenceModel   { return s.UserResidenceModel.(*mockResidenceModel) }

// createTestUser inserts a user into the mock store.
func createTestUser(t *testing.T, m *mockUserBaseModel, id int64, phone string) *model.UserBase {
	t.Helper()
	u := &model.UserBase{
		Id: id, Phone: phone, Nickname: sql.NullString{String: "测试用户", Valid: true},
		Status: model.UserStatusActive, CreditScore: 100, CreatedTime: time.Now(), UpdatedTime: time.Now(),
	}
	m.data[u.Id] = u
	m.phone[u.Phone] = u
	return u
}

// createTestMembership inserts a membership into the mock store.
func createTestMembership(t *testing.T, m *mockMembershipModel, id, uid, cid int64) *model.UserCommunityMembership {
	t.Helper()
	ms := &model.UserCommunityMembership{
		Id: id, UserId: uid, CommunityId: cid, BindStatus: model.MembershipBindStatusActive,
		JoinTime: time.Now(), CreatedTime: time.Now(), UpdatedTime: time.Now(),
	}
	m.data[ms.Id] = ms
	m.byUserCommIdx[fmt.Sprintf("%d_%d", uid, cid)] = ms.Id
	return ms
}

// createTestRole inserts a role into the mock store.
func createTestRole(t *testing.T, m *mockRoleModel, id, uid, mid, cid int64, roleCode string, verfStatus int64) *model.UserMembershipRole {
	t.Helper()
	r := &model.UserMembershipRole{
		Id: id, UserId: uid, MembershipId: sql.NullInt64{Int64: mid, Valid: true},
		CommunityId: cid, RoleCode: roleCode, VerfStatus: verfStatus,
		CreatedTime: time.Now(), UpdatedTime: time.Now(),
	}
	m.data[r.Id] = r
	return r
}
