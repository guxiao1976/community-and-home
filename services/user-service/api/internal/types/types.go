package types

// =============================================================================
// 用户管理
// =============================================================================

type ListUsersReq struct {
	Page     int32  `form:"page,optional,default=1"`
	PageSize int32  `form:"page_size,optional,default=10"`
	Keyword  string `form:"keyword,optional"`
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
	Id          int64  `json:"id,string"`
	UserId      int64  `json:"user_id,string"`
	CommunityId int64  `json:"community_id,string"`
	BindStatus  int32  `json:"bind_status"`
	JoinTime    int64  `json:"join_time"`
	LeaveTime   int64  `json:"leave_time"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}
