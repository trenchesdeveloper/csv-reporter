package main

import (
	"context"
	"github.com/google/uuid"
	db "github.com/trenchesdeveloper/csv-reporter/db/sqlc"
	"github.com/trenchesdeveloper/csv-reporter/helpers"
	"log"
	"net/http"
	"strings"
)

func NewAuthMiddleware(jwtManager *helpers.JwtManager, store db.Store) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the token from the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errorResponse(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}

			if parts := strings.Split(authHeader, "Bearer "); len(parts) != 2 {
				errorResponse(w, http.StatusUnauthorized, "Invalid authorization")
				return
			} else {
				authHeader = strings.TrimSpace(parts[1])
			}

			// Verify the token
			claims, err := jwtManager.ValidateToken(authHeader)
			if err != nil {
				log.Printf("Error validating token: %v", err)
				errorResponse(w, http.StatusUnauthorized, "UnAuthorized")
				return
			}

			if !jwtManager.IsAccessToken(claims) {
				errorResponse(w, http.StatusUnauthorized, "Not an access token")
				return
			}

			// extract the user ID from the claims
			userIdStr, err := claims.Claims.GetSubject()
			if err != nil {
				errorResponse(w, http.StatusUnauthorized, "UnAuthorized")
				return
			}

			// check if the user ID is valid uuid
			userId, err := uuid.Parse(userIdStr)
			if err != nil {
				errorResponse(w, http.StatusUnauthorized, "UnAuthorized")
				return
			}

			// Check if the user exists in the database
			user, err := store.FindUserById(r.Context(), userId)
			if err != nil {
				errorResponse(w, http.StatusUnauthorized, "User not found")
				return
			}

			// Set the user in the request context
			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(r *http.Request) (db.User, bool) {
	user, ok := r.Context().Value("user").(db.User)
	if !ok {
		return db.User{}, false
	}
	return user, true
}