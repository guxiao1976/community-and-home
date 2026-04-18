package middleware

import (
	"net/http"

	"github.com/guxiao/community-and-home/common/responsex"
)

type ScopeMiddleware struct {
}

func NewScopeMiddleware() *ScopeMiddleware {
	return &ScopeMiddleware{}
}

func (m *ScopeMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user scope from context (set by JWT middleware)
		scopeId := r.Context().Value("scopeId")

		// Headquarters users (scopeId = nil or 0) can access all data
		if scopeId == nil {
			next(w, r)
			return
		}

		// For non-headquarters users, validate scope access
		// Extract target scope from request (query param or path param)
		targetScope := r.URL.Query().Get("scope_id")
		if targetScope == "" {
			// No specific scope requested, allow access
			next(w, r)
			return
		}

		// TODO: Validate that user's scope includes target scope
		// This requires checking the administrative division hierarchy
		// For now, allow all requests

		next(w, r)
	}
}
