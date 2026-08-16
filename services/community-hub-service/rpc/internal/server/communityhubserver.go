package server

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/contact"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/lostfound"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/notice"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
)

// ContentPostServiceServer 通用图文发布 gRPC Server（原 NoticeServiceServer 改名，D4）。
type ContentPostServiceServer struct {
	svcCtx *svc.ServiceContext
	communityv1.UnimplementedContentPostServiceServer
}

func NewContentPostServiceServer(svcCtx *svc.ServiceContext) *ContentPostServiceServer {
	return &ContentPostServiceServer{svcCtx: svcCtx}
}

func (s *ContentPostServiceServer) CreateContentPost(ctx context.Context, in *communityv1.CreateContentPostRequest) (*communityv1.CreateContentPostResponse, error) {
	return notice.NewCreateContentPostLogic(ctx, s.svcCtx).CreateContentPost(in)
}

func (s *ContentPostServiceServer) ListContentPosts(ctx context.Context, in *communityv1.ListContentPostsRequest) (*communityv1.ListContentPostsResponse, error) {
	return notice.NewListContentPostsLogic(ctx, s.svcCtx).ListContentPosts(in)
}

func (s *ContentPostServiceServer) GetContentPost(ctx context.Context, in *communityv1.GetContentPostRequest) (*communityv1.GetContentPostResponse, error) {
	return notice.NewGetContentPostLogic(ctx, s.svcCtx).GetContentPost(in)
}

func (s *ContentPostServiceServer) UpdateContentPost(ctx context.Context, in *communityv1.UpdateContentPostRequest) (*communityv1.UpdateContentPostResponse, error) {
	return notice.NewUpdateContentPostLogic(ctx, s.svcCtx).UpdateContentPost(in)
}

func (s *ContentPostServiceServer) DeleteContentPost(ctx context.Context, in *communityv1.DeleteContentPostRequest) (*communityv1.DeleteContentPostResponse, error) {
	return notice.NewDeleteContentPostLogic(ctx, s.svcCtx).DeleteContentPost(in)
}

func (s *ContentPostServiceServer) GetPublishPermission(ctx context.Context, in *communityv1.GetPublishPermissionRequest) (*communityv1.GetPublishPermissionResponse, error) {
	return notice.NewGetPublishPermissionLogic(ctx, s.svcCtx).GetPublishPermission(in)
}

func (s *ContentPostServiceServer) GetMarqueeNotices(ctx context.Context, in *communityv1.GetMarqueeNoticesRequest) (*communityv1.GetMarqueeNoticesResponse, error) {
	return notice.NewGetMarqueeNoticesLogic(ctx, s.svcCtx).GetMarqueeNotices(in)
}

// ContactServiceServer 便民联络 gRPC Server
type ContactServiceServer struct {
	svcCtx *svc.ServiceContext
	communityv1.UnimplementedContactServiceServer
}

func NewContactServiceServer(svcCtx *svc.ServiceContext) *ContactServiceServer {
	return &ContactServiceServer{svcCtx: svcCtx}
}

func (s *ContactServiceServer) ListContacts(ctx context.Context, in *communityv1.ListContactsRequest) (*communityv1.ListContactsResponse, error) {
	return contact.NewListContactsLogic(ctx, s.svcCtx).ListContacts(in)
}

func (s *ContactServiceServer) UpsertContacts(ctx context.Context, in *communityv1.UpsertContactsRequest) (*communityv1.UpsertContactsResponse, error) {
	return contact.NewUpsertContactsLogic(ctx, s.svcCtx).UpsertContacts(in)
}

// LostFoundServiceServer 寻失互助 gRPC Server
type LostFoundServiceServer struct {
	svcCtx *svc.ServiceContext
	communityv1.UnimplementedLostFoundServiceServer
}

func NewLostFoundServiceServer(svcCtx *svc.ServiceContext) *LostFoundServiceServer {
	return &LostFoundServiceServer{svcCtx: svcCtx}
}

func (s *LostFoundServiceServer) CreateLostFound(ctx context.Context, in *communityv1.CreateLostFoundRequest) (*communityv1.CreateLostFoundResponse, error) {
	return lostfound.NewCreateLostFoundLogic(ctx, s.svcCtx).CreateLostFound(in)
}

func (s *LostFoundServiceServer) ListLostFound(ctx context.Context, in *communityv1.ListLostFoundRequest) (*communityv1.ListLostFoundResponse, error) {
	return lostfound.NewListLostFoundLogic(ctx, s.svcCtx).ListLostFound(in)
}

func (s *LostFoundServiceServer) GetLostFound(ctx context.Context, in *communityv1.GetLostFoundRequest) (*communityv1.GetLostFoundResponse, error) {
	return lostfound.NewGetLostFoundLogic(ctx, s.svcCtx).GetLostFound(in)
}

func (s *LostFoundServiceServer) ResolveLostFound(ctx context.Context, in *communityv1.ResolveLostFoundRequest) (*communityv1.ResolveLostFoundResponse, error) {
	return lostfound.NewResolveLostFoundLogic(ctx, s.svcCtx).ResolveLostFound(in)
}

func (s *LostFoundServiceServer) UpdateLostFoundModerationStatus(ctx context.Context, in *communityv1.UpdateModerationStatusRequest) (*communityv1.UpdateModerationStatusResponse, error) {
	return lostfound.NewUpdateLostFoundModerationStatusLogic(ctx, s.svcCtx).UpdateLostFoundModerationStatus(in)
}
