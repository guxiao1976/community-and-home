package user

import (
	"database/sql"
	"time"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	"github.com/guxiao1976/community-common/v2/pkg/crypto"
	"github.com/guxiao1976/community-user/model"
)

// ==================== Proto Conversion Helpers ====================

// int32Ptr 返回 int32 指针
func int32Ptr(v int32) *int32 {
	return &v
}

// int64Ptr 返回 int64 指针
func int64Ptr(v int64) *int64 {
	return &v
}

// defaultExpiryTime 返回角色的默认过期时间（无 sysconfig 时用硬编码）
func defaultExpiryTime(roleCode string) sql.NullTime {
	defaults := map[string]int64{
		"grid_worker":     8760,
		"community_admin": 17520,
		"committee":       17520,
		"property_admin":  8760,
		"tenant":          8760,
	}
	if hours, ok := defaults[roleCode]; ok {
		return sql.NullTime{Time: time.Now().Add(time.Duration(hours) * time.Hour), Valid: true}
	}
	return sql.NullTime{}
}

// toProtoUser converts a model.UserBase to a proto User.
func toProtoUser(u *model.UserBase) *userv1.User {
	if u == nil {
		return nil
	}
	phone := u.Phone
	if decrypted, err := crypto.AESDecrypt(phone); err == nil {
		phone = decrypted
	}
	user := &userv1.User{
		Id:           u.Id,
		Phone:        phone,
		AvatarUrl:    u.AvatarUrl.String,
		RealName:     u.RealName.String,
		IdCardNumber: u.IdCardNumber.String,
		Status:       int32(u.Status),
		CreditScore:  int32(u.CreditScore),
		Preferences:  u.Preferences.String,
		CreatedAt:    u.CreatedTime.Unix(),
		UpdatedAt:    u.UpdatedTime.Unix(),
	}
	if u.Nickname.Valid {
		user.Nickname = u.Nickname.String
	}
	if u.Gender.Valid {
		user.Gender = int32(u.Gender.Int64)
	}
	if u.BirthDate.Valid {
		user.BirthDate = u.BirthDate.Time.Format("2006-01-02")
	}
	return user
}

func toProtoUsers(users []*model.UserBase) []*userv1.User {
	result := make([]*userv1.User, 0, len(users))
	for _, u := range users {
		result = append(result, toProtoUser(u))
	}
	return result
}

func toProtoMembership(m *model.UserCommunityMembership) *userv1.CommunityMembership {
	if m == nil {
		return nil
	}
	cm := &userv1.CommunityMembership{
		Id:          m.Id,
		UserId:      m.UserId,
		CommunityId: m.CommunityId,
		BindStatus:  int32(m.BindStatus),
		JoinTime:    m.JoinTime.Unix(),
		CreatedAt:   m.CreatedTime.Unix(),
		UpdatedAt:   m.UpdatedTime.Unix(),
		Building:    int32(m.Building),
		Unit:        int32(m.Unit),
		Room:        int32(m.Room),
	}
	if m.LeaveTime.Valid {
		cm.LeaveTime = m.LeaveTime.Time.Unix()
	}
	return cm
}

func toProtoMemberships(memberships []*model.UserCommunityMembership) []*userv1.CommunityMembership {
	result := make([]*userv1.CommunityMembership, 0, len(memberships))
	for _, m := range memberships {
		result = append(result, toProtoMembership(m))
	}
	return result
}

func toProtoCertification(c *model.UserCertification) *userv1.Certification {
	if c == nil {
		return nil
	}
	cert := &userv1.Certification{
		Id:           c.Id,
		RoleId:       c.RoleId,
		UserId:       c.UserId,
		DocumentUrls: c.DocumentUrls.String,
		Status:       int32(c.Status),
		SubmitTime:   c.SubmitTime.Unix(),
	}
	if c.ReviewerId.Valid {
		cert.ReviewerId = c.ReviewerId.Int64
	}
	if c.ReviewNotes.Valid {
		cert.ReviewNotes = c.ReviewNotes.String
	}
	if c.ReviewTime.Valid {
		cert.ReviewTime = c.ReviewTime.Time.Unix()
	}
	return cert
}

func toProtoCertifications(certs []*model.UserCertification) []*userv1.Certification {
	result := make([]*userv1.Certification, 0, len(certs))
	for _, c := range certs {
		result = append(result, toProtoCertification(c))
	}
	return result
}

func toProtoResidence(r *model.UserResidence) *userv1.Residence {
	if r == nil {
		return nil
	}
	res := &userv1.Residence{
		Id:           r.Id,
		MembershipId: r.MembershipId,
		HouseId:      r.HouseId,
		Building:     r.Building,
		Unit:         r.Unit,
		Room:         r.Room,
		IsPrimary:    int32(r.IsPrimary),
		CreatedAt:    r.CreatedTime.Unix(),
		UpdatedAt:    r.UpdatedTime.Unix(),
	}
	if r.StartDate.Valid {
		res.StartDate = r.StartDate.Time.Format("2006-01-02")
	}
	if r.EndDate.Valid {
		res.EndDate = r.EndDate.Time.Format("2006-01-02")
	}
	return res
}

// ==================== Utility Helpers ====================

// buildHouseId constructs house_id from building, unit, room.
// Format: "building-unit-room" or "building-room" if unit is empty.
func buildHouseId(building, unit, room string) string {
	if unit == "" {
		return building + "-" + room
	}
	return building + "-" + unit + "-" + room
}

// parseDate parses a YYYY-MM-DD string to sql.NullTime.
func parseDate(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: t, Valid: true}
}
