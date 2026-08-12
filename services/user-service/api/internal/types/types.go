package types

// =============================================================================
// 用户管理
// =============================================================================

type ListUsersReq struct {
	Page     int32  `form:"page,optional,default=1"`
	PageSize int32  `form:"page_size,optional,default=10"`
	Keyword  string `form:"keyword,optional"`
	Phone    string `form:"phone,optional"`
	Status   *int32 `form:"status,optional"`
}

type ListUsersResp struct {
	Users []UserInfo `json:"users"`
	Page  PageInfo   `json:"page"`
}

type CreateUserReq struct {
	Phone    string `json:"phone"`
	Nickname string `json:"nickname,optional"`
}

type CreateUserResp struct {
	UserId int64 `json:"user_id,string"`
}

type GetUserReq struct {
	Id int64 `path:"id"`
}

type GetUserResp struct {
	User UserInfo `json:"user"`
}

type UpdateUserReq struct {
	Id        int64   `path:"id"`
	Nickname  *string `json:"nickname,optional"`
	AvatarUrl *string `json:"avatar_url,optional"`
	Status    *int32  `json:"status,optional"`
	Gender    *int32  `json:"gender,optional"`
	BirthDate *string `json:"birth_date,optional"`
}

type UpdateUserResp struct {
	User UserInfo `json:"user"`
}

type DeleteUserReq struct {
	Id int64 `path:"id"`
}

type DeleteUserResp struct {
	Success bool `json:"success"`
}

// =============================================================================
// 通用类型
// =============================================================================

type UserInfo struct {
	Id           int64  `json:"id,string"`
	Phone        string `json:"phone"`
	Nickname     string `json:"nickname"`
	AvatarUrl    string `json:"avatar_url"`
	RealName     string `json:"real_name"`
	IdCardNumber string `json:"id_card_number"`
	Gender       int32  `json:"gender"`
	BirthDate    string `json:"birth_date"`
	Status       int32  `json:"status"`
	CreditScore  int32  `json:"credit_score"`
	Preferences  string `json:"preferences"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

type PageInfo struct {
	Page       int32 `json:"page"`
	PageSize   int32 `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int32 `json:"total_pages"`
}

// =============================================================================
// 社区成员
// =============================================================================

type JoinCommunityReq struct {
	CommunityId int64 `json:"community_id,string"`
	Building    int32 `json:"building,optional"`
	Unit        int32 `json:"unit,optional"`
	Room        int32 `json:"room,optional"`
	// Ownership 权属：1=自有(owner) 2=租住(tenant)，必填（permission 数据权限自动授权）
	Ownership int32 `json:"ownership,optional"`
}

type JoinCommunityResp struct {
	Membership CommunityMembership `json:"membership"`
}

type GetMembershipsReq struct{}

type GetMembershipsResp struct {
	Memberships []CommunityMembership `json:"memberships"`
}

type LeaveCommunityReq struct {
	CommunityId int64 `json:"community_id,string"`
}

type LeaveCommunityResp struct{}

type CommunityMembership struct {
	Id          int64 `json:"id,string"`
	UserId      int64 `json:"user_id,string"`
	CommunityId int64 `json:"community_id,string"`
	BindStatus  int32 `json:"bind_status"`
	JoinTime    int64 `json:"join_time"`
	LeaveTime   int64 `json:"leave_time"`
	CreatedAt   int64 `json:"created_at"`
	UpdatedAt   int64 `json:"updated_at"`
	Building    int   `json:"building"` // 楼号
	Unit        int   `json:"unit"`     // 单元号
	Room        int   `json:"room"`     // 房号
}

// =============================================================================
// 房产绑定
// =============================================================================

type BindResidenceReq struct {
	MembershipId int64  `json:"membership_id,string"`
	Building     string `json:"building"`
	Unit         string `json:"unit"`
	Room         string `json:"room"`
	IsPrimary    int32  `json:"is_primary,optional"`
	StartDate    string `json:"start_date,optional"`
	EndDate      string `json:"end_date,optional"`
}

type BindResidenceResp struct {
	Residence Residence `json:"residence"`
}

type Residence struct {
	Id           int64  `json:"id,string"`
	MembershipId int64  `json:"membership_id,string"`
	HouseId      string `json:"house_id"`
	Building     string `json:"building"`
	Unit         string `json:"unit"`
	Room         string `json:"room"`
	IsPrimary    int32  `json:"is_primary"`
	StartDate    string `json:"start_date,optional"`
	EndDate      string `json:"end_date,optional"`
}

// =============================================================================
// 角色申请
// =============================================================================

type ApplyRoleReq struct {
	CommunityId int64  `json:"community_id,string"`
	RoleCode    string `json:"role_code"`
}

type ApplyRoleResp struct{}

type GetUserRolesResp struct {
	Roles []RoleInfo `json:"roles"`
}

type RoleInfo struct {
	Id          int64  `json:"id,string"`
	UserId      int64  `json:"user_id,string"`
	CommunityId int64  `json:"community_id,string"`
	RoleCode    string `json:"role_code"`
	VerfStatus  int32  `json:"verf_status"`
}

// =============================================================================
// 认证（Certification）
// =============================================================================

type SubmitCertificationReq struct {
	RoleId       int64    `json:"role_id,string"`
	DocumentUrls []string `json:"document_urls"`
	RealName     string   `json:"real_name"`
	IdCardNumber string   `json:"id_card_number"`
	Building     string   `json:"building,optional"`
	Unit         string   `json:"unit,optional"`
	Room         string   `json:"room,optional"`
}

type SubmitCertificationResp struct {
	Certification CertificationInfo `json:"certification"`
}

type ReviewCertificationReq struct {
	Result      int32  `json:"result"`
	ReviewNotes string `json:"review_notes,optional"`
	ExpiresAt   string `json:"expires_at,optional"`
}

type ReviewCertificationResp struct{}

type CertificationInfo struct {
	Id           int64  `json:"id,string"`
	RoleId       int64  `json:"role_id,string"`
	UserId       int64  `json:"user_id,string"`
	RoleCode     string `json:"role_code,optional"`
	DocumentUrls string `json:"document_urls"`
	RealName     string `json:"real_name,optional"`
	Status       int32  `json:"status"`
	ReviewerId   int64  `json:"reviewer_id,string,optional"`
	ReviewNotes  string `json:"review_notes,optional"`
	ReviewTime   int64  `json:"review_time,string,optional"`
	SubmitTime   int64  `json:"submit_time,string"`
}

type ListCertificationsReq struct {
	Page     int32  `form:"page,optional,default=1"`
	PageSize int32  `form:"page_size,optional,default=10"`
	Status   *int32 `form:"status,optional"`
}

type ListCertificationsResp struct {
	Certifications []CertificationInfo `json:"certifications"`
	Total          int64               `json:"total,string"`
}

type GetMyCertificationsReq struct {
	UserId int64 `path:"userId"`
}

type GetMyCertificationsResp struct {
	Certifications []CertificationInfo `json:"certifications"`
}
