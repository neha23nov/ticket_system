// Package middleware provides HTTP middleware functions that run before
// the main request handler. Middleware is like a security checkpoint —
// it intercepts every request, performs a check, and either allows the
// request to continue or rejects it.
//
// This file contains the JWT authentication middleware. It protects
// endpoints that require a logged-in user by checking for a valid
// JWT token in the request headers.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"ticket-system/internal/auth"
)

// contextKey is a custom type used for storing values in the request context.
// Using a custom type (instead of a plain string) avoids accidental collisions
// with context keys from other packages.
type contextKey string

// UserIDKey is the key used to store and retrieve the user's ID from the
// request context. After the JWT middleware validates a token, it stores
// the user's ID under this key so that handlers can access it.
const UserIDKey contextKey = "userID"

// JWTAuth returns a middleware function that protects routes by requiring
// a valid JWT token in the request's Authorization header.
//
// How it works:
//  1. Looks for the "Authorization" header in the request.
//  2. Checks that the header value starts with "Bearer " followed by the token.
//  3. Validates the token using the provided secret key.
//  4. If valid, stores the user's ID in the request context and allows the
//     request to continue to the actual handler.
//  5. If invalid, returns a 401 Unauthorized response immediately.
//
// Parameters:
//   - secret: The JWT secret key used to validate tokens.
//
// Usage (with chi router):
//
//	r.Group(func(r chi.Router) {
//	    r.Use(middleware.JWTAuth("my-secret-key"))
//	    r.Get("/tickets", ticketHandler.List)
//	})
func JWTAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Step 1: Get the Authorization header from the request.
			// Expected format: "Bearer eyJhbGciOiJIUzI1NiIs..."
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"authorization header is required"}`, http.StatusUnauthorized)
				return
			}

			// Step 2: Split the header into "Bearer" and the actual token.
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error":"authorization header must be in format: Bearer <token>"}`, http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			// Step 3: Validate the token and extract the user ID.
			userID, err := auth.ParseToken(tokenString, secret)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Step 4: Store the user ID in the request context.
			// This makes the user ID available to all downstream handlers
			// via: userID := r.Context().Value(middleware.UserIDKey).(uint)
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID is a helper function that extracts the authenticated user's ID
// from the request context. It should only be called in handlers that are
// protected by the JWTAuth middleware.
//
// Returns the user ID and true if found, or 0 and false if not found
// (which would indicate the middleware was not applied to this route).
func GetUserID(r *http.Request) (uint, bool) {
	userID, ok := r.Context().Value(UserIDKey).(uint)
	return userID, ok
}
