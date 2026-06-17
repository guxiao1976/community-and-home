package types

// =============================================================================
// 通知公告
// =============================================================================

type CreateNoticeReq struct {
	CommunityId int64  `json:"community_id,string"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Role        int32  `json:"role"`
	Publisher   string `json:"publisher"`
	PublisherId int64  `json:"publisher_id,string"`
}

type CreateNoticeResp struct {
	Id int64 `json:"id,string"`
}

type ListNoticesReq struct {
	CommunityId int64 `form:"community_id"`
	Role        int32 `form:"role,optional"`
	Page        int32 `form:"page,optional,default=1"`
	PageSize    int32 `form:"page_size,optional,default=10"`
}

type ListNoticesResp struct {
	Notices []NoticeInfo `json:"notices"`
	Total   int64        `json:"total,string"`
}

type GetNoticeReq struct {
	Id int64 `path:"id"`
}

type GetNoticeResp struct {
	Notice NoticeInfo `json:"notice"`
}

type UpdateNoticeReq struct {
	Id       int64  `path:"id"`
	Title    string `json:"title,optional"`
	Content  string `json:"content,optional"`
	IsPinned bool   `json:"is_pinned,optional"`
}

type DeleteNoticeReq struct {
	Id int64 `path:"id"`
}

// =============================================================================
// 便民联络
// =============================================================================

type ListContactsReq struct {
	CommunityId int64 `form:"community_id"`
}

type ListContactsResp struct {
	Contacts []ContactInfo `json:"contacts"`
}

type ContactEntry struct {
	Category int32  `json:"category"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
}

type UpsertContactsReq struct {
	CommunityId int64          `json:"community_id,string"`
	Contacts    []ContactEntry `json:"contacts"`
}

// =============================================================================
// 寻失互助
// =============================================================================

type CreateLostFoundReq struct {
	CommunityId  int64    `json:"community_id,string"`
	Type         int32    `json:"type"`
	Title        string   `json:"title"`
	Description  string   `json:"description,optional"`
	ImageUrls    []string `json:"image_urls,optional"`
	ContactPhone string   `json:"contact_phone,optional"`
	PublisherId  int64    `json:"publisher_id,string"`
}

type CreateLostFoundResp struct {
	Id int64 `json:"id,string"`
}

type ListLostFoundReq struct {
	CommunityId int64 `form:"community_id"`
	Type        int32 `form:"type,optional"`
	Page        int32 `form:"page,optional,default=1"`
	PageSize    int32 `form:"page_size,optional,default=10"`
}

type ListLostFoundResp struct {
	Items []LostFoundItemInfo `json:"items"`
	Total int64               `json:"total,string"`
}

type GetLostFoundReq struct {
	Id int64 `path:"id"`
}

type GetLostFoundResp struct {
	Item LostFoundItemInfo `json:"item"`
}

type ResolveLostFoundReq struct {
	Id int64 `path:"id"`
}

// =============================================================================
// 通用类型
// =============================================================================

type NoticeInfo struct {
	Id          int64                      `json:"id,string"`
	CommunityId int64                      `json:"community_id,string"`
	Title       string                     `json:"title"`
	Content     string                     `json:"content"`
	Role        int32                      `json:"role"`
	Publisher   string                     `json:"publisher"`
	PublisherId int64                      `json:"publisher_id,string"`
	IsPinned    bool                       `json:"is_pinned"`
	PublishedAt int64                      `json:"published_at"`
	CreatedAt   int64                      `json:"created_at"`
	UpdatedAt   int64                      `json:"updated_at"`
	Attachments []NoticeAttachmentInfo     `json:"attachments"`
}

type NoticeAttachmentInfo struct {
	Id       int64  `json:"id,string"`
	FileName string `json:"file_name"`
	FileUrl  string `json:"file_url"`
	FileSize int64  `json:"file_size,string"`
}

type ContactInfo struct {
	Id          int64  `json:"id,string"`
	CommunityId int64  `json:"community_id,string"`
	Category    int32  `json:"category"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	SortOrder   int32  `json:"sort_order"`
}

type LostFoundItemInfo struct {
	Id           int64    `json:"id,string"`
	CommunityId  int64    `json:"community_id,string"`
	Type         int32    `json:"type"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	ImageUrls    []string `json:"image_urls"`
	ContactPhone string   `json:"contact_phone"`
	Status       int32    `json:"status"`
	PublisherId  int64    `json:"publisher_id,string"`
	CreatedAt    int64    `json:"created_at"`
}
