package httpapi

import (
	"context"
	"net/http"
	"strings"

	"tarkovhelper-api/internal/security"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromCtx(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDKey)
	s, ok := v.(string)
	return s, ok && s != ""
}

func authMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or missing token")
				return
			}
			token := strings.TrimSpace(h[len("Bearer "):])
			userID, err := security.ParseHS256(token, jwtSecret)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or missing token")
				return
			}
			next.ServeHTTP(w, r.WithContext(withUserID(r.Context(), userID)))
		})
	}
}
