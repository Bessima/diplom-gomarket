package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Bessima/diplom-gomarket/internal/handlers"
)

func AuthMiddleware(authHandler *handlers.AuthHandler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			cookie, err := r.Cookie("access_token")
			if err == nil {
				tokenString = cookie.Value
			} else {
				authHeader := r.Header.Get("Authorization")
				if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
					tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if tokenString == "" {
				http.Error(w, "Authorization token required", http.StatusUnauthorized)
				return
			}

			claims, err := authHandler.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}
			user, err := authHandler.UserStorage.GetUserByID(claims.UserID)
			if err != nil || user == nil {
				http.Error(w, "User not found", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), handlers.UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
