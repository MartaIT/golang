package middleware

import (
	"context"
	"net/http"
	"strings"

	"golang/utils"
)

type ctxKey string

const (
	UserCtxKey ctxKey = "user"
)

// AuthMiddleware memeriksa header Authorization: Bearer <token>
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing Authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		token := parts[1]
		claims, err := utils.ParseToken(token)
		if err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}
		// simpan claims di context
		ctx := context.WithValue(r.Context(), UserCtxKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware generator untuk role-based access
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := r.Context().Value(UserCtxKey)
			if v == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			claims, ok := v.(*utils.Claims)
			if !ok {
				http.Error(w, "invalid auth context", http.StatusUnauthorized)
				return
			}
			if claims.Role != role {
				http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
