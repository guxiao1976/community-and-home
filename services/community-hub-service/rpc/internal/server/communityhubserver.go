package server

import (
	"context"

	communityv1 "github.com/guxiao1976/api-proto/gen/go/community/v1"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/contact"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/lostfound"
	"github.com/guxiao1976/community-hub/rpc/internal/logic/notice"
	"github.com/guxiao1976/community-hub/rpc/internal/svc"
)

// NoticeServiceServer 通知公告 gRPC Server
type NoticeServiceServer struct {
	svcCtx *svc.ServiceContext
	communityv1.UnimplementedNoticeServiceServer
}

func NewNoticeServiceServer(svcCtx *svc.ServiceContext) *NoticeServiceServer {
	return &NoticeServiceServer{svcCtx: svcCtx}
}

func (s *NoticeServiceServer) CreateNotice(ctx context.Context, in *communityv1.CreateNoticeRequest) (*communityv1.CreateNoticeResponse, error) {
	return notice.NewCreateNoticeLogic(ctx, s.svcCtx).CreateNotice(in)
}

func (s *NoticeServiceServer) ListNotices(ctx context.Context, in *communityv1.ListNoticesRequest) (*communityv1.ListNoticesResponse, error) {
	return notice.NewListNoticesLogic(ctx, s.svcCtx).ListNotices(in)
}

func (s *NoticeServiceServer) GetNotice(ctx context.Context, in *communityv1.GetNoticeRequest) (*communityv1.GetNoticeResponse, error) {
	return notice.NewGetNoticeLogic(ctx, s.svcCtx).GetNotice(in)
}

func (s *NoticeServiceServer) UpdateNotice(ctx context.Context, in *communityv1.UpdateNoticeRequest) (*communityv1.UpdateNoticeResponse, error) {
	return notice.NewUpdateNoticeLogic(ctx, s.svcCtx).UpdateNotice(in)
}

func (s *NoticeServiceServer) DeleteNotice(ctx context.Context, in *communityv1.DeleteNoticeRequest) (*communityv1.DeleteNoticeResponse, error) {
	return notice.NewDeleteNoticeLogic(ctx, s.svcCtx).DeleteNotice(in)
}

func (s *NoticeServiceServer) UpdateNoticeModerationStatus(ctx context.Context, in *communityv1.UpdateModerationStatusRequest) (*communityv1.UpdateModerationStatusResponse, error) {
	return notice.NewUpdateNoticeModerationStatusLogic(ctx, s.svcCtx).UpdateNoticeModerationStatus(in)
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
