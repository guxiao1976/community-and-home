package server

import (
	"context"

	filev1 "github.com/guxiao1976/api-proto/gen/go/file/v1"
	"github.com/guxiao1976/community-file/rpc/internal/logic/file"
	"github.com/guxiao1976/community-file/rpc/internal/svc"
)

type FileServiceServer struct {
	svcCtx *svc.ServiceContext
	filev1.UnimplementedFileServiceServer
}

func NewFileServiceServer(svcCtx *svc.ServiceContext) *FileServiceServer {
	return &FileServiceServer{svcCtx: svcCtx}
}

func (s *FileServiceServer) GetUploadUrl(ctx context.Context, in *filev1.GetUploadUrlRequest) (*filev1.GetUploadUrlResponse, error) {
	l := file.NewGetUploadUrlLogic(ctx, s.svcCtx)
	return l.GetUploadUrl(in)
}

func (s *FileServiceServer) ConfirmUpload(ctx context.Context, in *filev1.ConfirmUploadRequest) (*filev1.ConfirmUploadResponse, error) {
	l := file.NewConfirmUploadLogic(ctx, s.svcCtx)
	return l.ConfirmUpload(in)
}

func (s *FileServiceServer) GetFileUrl(ctx context.Context, in *filev1.GetFileUrlRequest) (*filev1.GetFileUrlResponse, error) {
	l := file.NewGetFileUrlLogic(ctx, s.svcCtx)
	return l.GetFileUrl(in)
}

func (s *FileServiceServer) DeleteFile(ctx context.Context, in *filev1.DeleteFileRequest) (*filev1.DeleteFileResponse, error) {
	l := file.NewDeleteFileLogic(ctx, s.svcCtx)
	return l.DeleteFile(in)
}

func (s *FileServiceServer) ListFiles(ctx context.Context, in *filev1.ListFilesRequest) (*filev1.ListFilesResponse, error) {
	l := file.NewListFilesLogic(ctx, s.svcCtx)
	return l.ListFiles(in)
}
