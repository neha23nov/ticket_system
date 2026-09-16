// Package auth provides helper functions for secure password handling
// and JWT token management used by the authentication system.
//
// This file handles JWT (JSON Web Token) creation and validation. JWTs are
// used to authenticate users after login — the server issues a token, and the
// client sends it back with every request to prove their identity.
//
// How it works:
//  1. User logs in with email + password.
//  2. Server verifies credentials and creates a signed JWT containing the user's ID.
//  3. Client stores the token and sends it in the "Authorization: Bearer <token>" header.
//  4. Server validates the token on each request to identify the user.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// tokenClaims defines the data stored inside the JWT token.
// It embeds jwt.RegisteredClaims which provides standard fields like expiry time.
type tokenClaims struct {
	UserID uint `json:"user_id"` // The ID of the authenticated user.
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a given user.
//
// Parameters:
//   - userID: The unique ID of the user to encode in the token.
//   - secret: The secret key used to sign the token (only the server knows this).
//   - expiry: How long the token should remain valid (e.g., 24 hours).
//
// Returns the signed token string (e.g., "eyJhbGciOiJIUzI1NiIs...") or an error.
func GenerateToken(userID uint, secret string, expiry time.Duration) (string, error) {
	// Build the claims (the data inside the token).
	claims := tokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt tells the token when to stop being valid.
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			// IssuedAt records when the token was created.
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create a new token using the HS256 signing algorithm and our claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key and return the final string.
	return token.SignedString([]byte(secret))
}

// ParseToken takes a token string and the secret key, validates the token,
// and extracts the user ID from it.
//
// This is used by the authentication middleware to identify which user is
// making a request.
//
// Returns the user ID if the token is valid, or an error if:
//   - The token has expired
//   - The token was tampered with
//   - The token format is invalid
func ParseToken(tokenString, secret string) (uint, error) {
	// Parse and validate the token.
	token, err := jwt.ParseWithClaims(tokenString, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// This function provides the secret key for verification.
		// It also ensures the signing method is what we expect (HMAC).
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return 0, err
	}

	// Extract our custom claims from the validated token.
	claims, ok := token.Claims.(*tokenClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	return claims.UserID, nil
}
