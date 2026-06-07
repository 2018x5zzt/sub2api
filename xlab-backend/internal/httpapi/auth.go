package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const tokenContextKey contextKey = "token"

type API struct {
	core CoreClient
}

func (a *API) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header is required")
			return
		}
		if _, err := a.core.CurrentUser(r.Context(), token); err != nil {
			writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), tokenContextKey, token)))
	})
}

func bearerToken(header string) string {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func tokenFromContext(ctx context.Context) string {
	value, _ := ctx.Value(tokenContextKey).(string)
	return value
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": data})
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"code": code, "message": message})
}
