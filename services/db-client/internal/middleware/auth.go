package middleware

import (
	"context"
	"db-client/internal/utils"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string
type M map[string]string

const userIDKey contextKey = "userID"

// Auth validates the Superbase-issued JWT.
// It stores the UUID in context.

func Auth(keyfunc jwt.Keyfunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if raw == "" {
				utils.WriteJSON(w, http.StatusUnauthorized, M{"error": "unauthorized"})
				return
			}
			tok, err := jwt.Parse(raw, keyfunc, jwt.WithAudience("authenticated"), jwt.WithExpirationRequired())
			if err != nil || !tok.Valid {
				utils.WriteJSON(w, http.StatusUnauthorized, M{"error": "unauthorized"})
				return
			}
			sub, err := tok.Claims.GetSubject()
			if err != nil || sub == "" {
				utils.WriteJSON(w, http.StatusUnauthorized, M{"error": "unauthorized"})
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, sub)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin queries profile.role and returns 403 if not admin. Must be chained after Auth.

func RequireAdmin(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := UserIDFromContext(r.Context())
			if userID == "" {
				utils.WriteJSON(w, http.StatusForbidden, M{"error": "forbidden"})
				return
			}
			var role string
			err := pool.QueryRow(r.Context(), "SELECT role FROM profile WHERE id = $1", userID).Scan(&role)
			if err != nil || role != "admin" {
				utils.WriteJSON(w, http.StatusForbidden, M{"error": "forbidden"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func UserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}
