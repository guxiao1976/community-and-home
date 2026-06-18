package file

import (
	"net/http"

	"github.com/guxiao1976/community-file/api/internal/logic/file"
	"github.com/guxiao1976/community-file/api/internal/middleware"
	"github.com/guxiao1976/community-file/api/internal/svc"
	"github.com/guxiao1976/community-file/api/internal/types"
	"github.com/guxiao1976/community-common/v2/pkg/responsex"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetUploadUrlHandler 获取预签名上传 URL
func GetUploadUrlHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetUploadUrlReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.CtxError(r.Context(), err)
			return
		}

		userId, err := middleware.GetUserIdFromContext(r)
		if err != nil {
			responsex.CtxError(r.Context(), err)
			return
		}

		l := file.NewGetUploadUrlLogic(r.Context(), svcCtx)
		resp, err := l.GetUploadUrl(userId, &req)
		if err != nil {
			responsex.CtxError(r.Context(), err)
		} else {
			responsex.CtxSuccess(r.Context(), resp)
		}
	}
}

// ConfirmUploadHandler 确认上传完成
func ConfirmUploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ConfirmUploadReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.CtxError(r.Context(), err)
			return
		}

		userId, err := middleware.GetUserIdFromContext(r)
		if err != nil {
			responsex.CtxError(r.Context(), err)
			return
		}

		l := file.NewConfirmUploadLogic(r.Context(), svcCtx)
		resp, err := l.ConfirmUpload(userId, &req)
		if err != nil {
			responsex.CtxError(r.Context(), err)
		} else {
			responsex.CtxSuccess(r.Context(), resp)
		}
	}
}

// GetFileUrlHandler 获取文件下载 URL
func GetFileUrlHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetFileUrlReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.CtxError(r.Context(), err)
			return
		}

		userId, err := middleware.GetUserIdFromContext(r)
		if err != nil {
			responsex.CtxError(r.Context(), err)
			return
		}

		l := file.NewGetFileUrlLogic(r.Context(), svcCtx)
		resp, err := l.GetFileUrl(userId, &req)
		if err != nil {
			responsex.CtxError(r.Context(), err)
		} else {
			responsex.CtxSuccess(r.Context(), resp)
		}
	}
}

// DeleteFileHandler 删除文件
func DeleteFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteFileReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.CtxError(r.Context(), err)
			return
		}

		userId, err := middleware.GetUserIdFromContext(r)
		if err != nil {
			responsex.CtxError(r.Context(), err)
			return
		}

		l := file.NewDeleteFileLogic(r.Context(), svcCtx)
		if err := l.DeleteFile(userId, &req); err != nil {
			responsex.CtxError(r.Context(), err)
		} else {
			responsex.CtxSuccess(r.Context(), nil)
		}
	}
}

// ListFilesHandler 分页查询文件列表
func ListFilesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ListFilesReq
		if err := httpx.Parse(r, &req); err != nil {
			responsex.CtxError(r.Context(), err)
			return
		}

		userId, err := middleware.GetUserIdFromContext(r)
		if err != nil {
			responsex.CtxError(r.Context(), err)
			return
		}

		l := file.NewListFilesLogic(r.Context(), svcCtx)
		resp, err := l.ListFiles(userId, &req)
		if err != nil {
			responsex.CtxError(r.Context(), err)
		} else {
			responsex.CtxSuccess(r.Context(), resp)
		}
	}
}
