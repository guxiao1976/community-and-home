package handler

import (
	"net/http"

	"github.com/guxiao1976/community-file/api/internal/handler/file"
	"github.com/guxiao1976/community-file/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				// 获取预签名上传 URL
				Method:  http.MethodPost,
				Path:    "/upload-url",
				Handler: file.GetUploadUrlHandler(serverCtx),
			},
			{
				// 确认上传完成
				Method:  http.MethodPost,
				Path:    "/confirm",
				Handler: file.ConfirmUploadHandler(serverCtx),
			},
			{
				// 分页查询文件列表
				Method:  http.MethodGet,
				Path:    "/",
				Handler: file.ListFilesHandler(serverCtx),
			},
			{
				// 获取文件下载 URL
				Method:  http.MethodGet,
				Path:    "/:id",
				Handler: file.GetFileUrlHandler(serverCtx),
			},
			{
				// 删除文件
				Method:  http.MethodDelete,
				Path:    "/:id",
				Handler: file.DeleteFileHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/files"),
	)
}
