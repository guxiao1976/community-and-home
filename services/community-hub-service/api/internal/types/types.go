package types

// =============================================================================
// 业务错误码（08xxxx 命名空间，access-data-permission 阶段④ 注册）
// =============================================================================
//
// 错误码分层语义：
//   - 080002 无发布权限（功能权限）：由 PermMiddleware → permission CheckPermission 产出，
//     与具体数据范围无关，表示角色/能力不足；
//   - 080006 目标小区超出发布者数据范围（数据权限）：由 rpc 层落库前 AssertPublishScope 产出，
//     permission 侧为 060007（RPC 面），community-hub 统一映射为 080006（API 面）。
//
// 功能权限先于数据权限执行：中间件链（WithJwt → PermMiddleware）先于 handler → rpc 逻辑。
// 常量定义在 rpc/internal/logic/scope.CodePublishScopeDenied（=80006），本文件仅登记语义。
const (
	CodeContentPostNotFound  = 80001 // 内容帖不存在
	CodePublishDenied        = 80002 // 无发布权限（功能权限，复用）/ 非帖作者（Update/Delete 作者归属校验）
	CodeOverLimit            = 80003 // 发布目标数量超限（复用）
	CodeLostFoundMiss        = 80004 // 寻失记录不存在（复用）
	CodeInvalidParam         = 80005 // 参数无效（复用）
	CodeScopeDenied          = 80006 // 目标小区超出发布者数据范围（数据权限，新增）
	CodeSectionQuotaExceeded = 80007 // 超出发布配额（板块配额，新增；常量实现在 rpc/internal/logic/scope.CodeSectionQuotaExceeded）
)

// =============================================================================
// 通用图文发布（R2 wire 兼容：REST 响应键保持 notices/notice/content，路径保持 /notices）
// =============================================================================

type CreateContentPostReq struct {
	SectionCode   string   `json:"section_code"`
	Title         string   `json:"title"`
	Text          string   `json:"content"`       // REST wire 键保持 content（R2；RPC/proto/DB 用 text）
	EntryStatus   int32    `json:"entry_status"`  // 0=draft 默认 / 1=submitted（与 proto int32 同号，无数值变换）
	CommunityIds  []string `json:"community_ids"` // REST string 形式 Snowflake ID（encoding/json ,string 不支持 slice）
	AttachmentIds []string `json:"attachment_ids"`
	IsPinned      bool     `json:"is_pinned,optional"`
}

type CreateContentPostResp struct {
	Id int64 `json:"id,string"`
}

type ListContentPostsReq struct {
	CommunityId int64  `form:"community_id"`
	Role        int32  `form:"role,optional"`
	SectionCode string `form:"section_code,optional"`
	Page        int32  `form:"page,optional,default=1"`
	PageSize    int32  `form:"page_size,optional,default=10"`
	SinceDays   int32  `form:"since_days,optional"` // 可选时间窗口（天，1..365；缺省 0 → RPC 缺省不过滤，REVISION r2-2）
}

// ListContentPostsResp R2 wire 兼容：JSON 键保持 notices（移动端 getNoticeList 读 res.notices）。
type ListContentPostsResp struct {
	Notices []ContentPostInfo `json:"notices"`
	Total   int64             `json:"total,string"`
}

type GetContentPostReq struct {
	Id          int64 `path:"id"`
	CommunityId int64 `form:"community_id,optional"` // GET query form 绑定；optional 供 R2 兼容回退（RPC 层仍必填）
}

// GetContentPostResp R2 wire 兼容：JSON 键保持 notice（移动端 getNoticeDetail 读 res.notice）。
type GetContentPostResp struct {
	Notice ContentPostInfo `json:"notice"`
}

// UpdateContentPostReq V5 presence 语义：Title/Text/SectionCode/IsPinned 为 pointer（nil=未携带），
// CommunityIds/AttachmentIds 为全量替换集 + HasScopeChange/HasAttachmentChange bool 标志
// （false=不改，true=全量替换，空集=清空/080005）。
type UpdateContentPostReq struct {
	Id                  int64    `path:"id"`
	Title               *string  `json:"title,optional"`
	Text                *string  `json:"content,optional"` // REST wire 键保持 content（R2）
	SectionCode         *string  `json:"section_code,optional"`
	CommunityIds        []string `json:"community_ids,optional"`
	HasScopeChange      bool     `json:"has_scope_change,optional"` // true=按 community_ids 全量替换（空集→080005）；false=不改
	AttachmentIds       []string `json:"attachment_ids,optional"`
	HasAttachmentChange bool     `json:"has_attachment_change,optional"` // true=按 attachment_ids 全量替换（空集=清空）；false=不改
	IsPinned            *bool    `json:"is_pinned,optional"`             // *true 置顶 / *false 取消置顶
	Status              int32    `json:"status,optional"`                // 0=无提交动作（编辑）/ 1=submit（与 proto 同号）
}

type DeleteContentPostReq struct {
	Id int64 `path:"id"`
}

// GetMarqueeNoticesReq 跑马灯（板块固定 notice，评审 INFO 1）。
type GetMarqueeNoticesReq struct {
	CommunityId int64 `form:"community_id"`
}

type GetMarqueeNoticesResp struct {
	Items []ContentPostMarqueeItemInfo `json:"items"`
}

type GetPublishPermissionReq struct{}

type GetPublishPermissionResp struct {
	CanPublish       bool    `json:"can_publish"`
	PublishableRoles []int32 `json:"publishable_roles"`
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

// ContentPostInfo 帖体（R2 wire 兼容：正文 JSON 键保持 content，Go 字段 Text 显式映射）。
type ContentPostInfo struct {
	Id              int64                       `json:"id,string"`
	CommunityId     int64                       `json:"community_id,string"`
	Title           string                      `json:"title"`
	Text            string                      `json:"content"` // proto/DB 用 text，REST wire 用 content（R2 分轨，ADR）
	Role            int32                       `json:"role"`
	Publisher       string                      `json:"publisher"`
	PublisherId     int64                       `json:"publisher_id,string"`
	IsPinned        bool                        `json:"is_pinned"`
	PublishedAt     int64                       `json:"published_at"`
	CreatedAt       int64                       `json:"created_at"`
	UpdatedAt       int64                       `json:"updated_at"`
	SectionCode     string                      `json:"section_code"`
	Status          int32                       `json:"status"`
	AttachmentCount int32                       `json:"attachment_count"`
	Attachments     []ContentPostAttachmentInfo `json:"attachments"`
}

// ContentPostAttachmentInfo 附件（新增 file_type/file_id/review_status 走新键 additive）。
type ContentPostAttachmentInfo struct {
	Id           int64  `json:"id,string"`
	FileName     string `json:"file_name"`
	FileUrl      string `json:"file_url"`
	FileSize     int64  `json:"file_size,string"`
	FileType     string `json:"file_type"`
	FileId       int64  `json:"file_id,string"`
	ReviewStatus int32  `json:"review_status"`
}

// ContentPostMarqueeItemInfo 跑马灯条目（id+title）。
type ContentPostMarqueeItemInfo struct {
	Id    int64  `json:"id,string"`
	Title string `json:"title"`
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
