package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/atul-aggarwaal/cloud-drive/internal/pkg/crypto"
)

// define a custom  type of type string to store User ID extracted from JWT Token
// This will help our "string" key from others using similar string "user-id" or "ctx_verified_user_id"
// Thus preventing accidental collission.
type contextKey string

// Define an context keu for variable UserIdKey
const UserIdKey contextKey = "ctx_verified_user_id"

func AuthInterceptor(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Read authorization header token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Access Denied: Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Parse "Bearer <token>" from Authorization header
		tokenParts := strings.Split(authHeader, " ")

		// Validate, If Authorization header has correct formatting for Bearer token
		if len(tokenParts) != 2 || !strings.EqualFold(tokenParts[0], "Bearer") {
			http.Error(w, "Access Denied : Malformed Authorization Header formatting", http.StatusUnauthorized)
			return
		}

		// Read Access token
		tokenString := tokenParts[1]

		// Validate authenticity of the token
		claims, err := crypto.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Access Denied: Expired or Corrupted access token", http.StatusUnauthorized)
			return
		}

		// Inject user id in context
		ctx := context.WithValue(r.Context(), UserIdKey, claims.UserID)

		//call next method if token is valid
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
