package middleware

import (
	"bytes"
	"database/sql"
	"io"
	"net/http"
	"time"

	"github.com/guxiao/community-and-home/services/masterdata/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuditLogMiddleware struct {
	AuditLogModel model.MdAuditLogModel
}

func NewAuditLogMiddleware(auditLogModel model.MdAuditLogModel) *AuditLogMiddleware {
	return &AuditLogMiddleware{
		AuditLogModel: auditLogModel,
	}
}

func (m *AuditLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only audit write operations
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
			next.ServeHTTP(w, r)
			return
		}

		// Read request body
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Execute handler
		rw := &responseWriter{ResponseWriter: w, body: &bytes.Buffer{}}
		next.ServeHTTP(rw, r)

		// Determine entity type from path
		entityType := "unknown"
		action := r.Method

		path := r.URL.Path
		if containsSubstr(path, "/divisions") {
			entityType = "md_administrative_division"
		} else if containsSubstr(path, "/communities") {
			entityType = "md_community"
		}

		// Write audit log asynchronously
		go func() {
			var oldVal, newVal sql.NullString
			if len(bodyBytes) > 0 {
				oldVal = sql.NullString{String: string(bodyBytes), Valid: true}
			}
			respStr := rw.body.String()
			if respStr != "" {
				newVal = sql.NullString{String: respStr, Valid: true}
			}

			auditLog := &model.MdAuditLog{
				UserId:      0, // TODO: 从JWT获取
				EntityType:  entityType,
				EntityId:    0,  // TODO: 解析URL获取实体ID
				Action:      action,
				OldValue:    oldVal,
				NewValue:    newVal,
				IpAddress:   getClientIP(r),
				UserAgent:   toNullString(r.UserAgent()),
				CreatedTime: time.Now(),
			}
			_, err := m.AuditLogModel.Insert(nil, auditLog)
			if err != nil {
				logx.Errorf("failed to write audit log: %v", err)
			}
		}()
	}
}

type responseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}