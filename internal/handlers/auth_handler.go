// Package handlers contains all the HTTP request handlers for the ticket system.
//
// This file handles user authentication — registration and login.
//
// Registration flow:
//  1. User sends their username, email, and password.
//  2. Server validates the input and checks for duplicates.
//  3. Password is hashed using bcrypt (never stored as plain text).
//  4. User record is created in the database.
//
// Login flow:
//  1. User sends their email and password.
//  2. Server looks up the user by email.
//  3. Server compares the provided password against the stored hash.
//  4. If correct, a JWT token is generated and returned.
//  5. The client uses this token for all subsequent authenticated requests.
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"ticket-system/internal/auth"
	"ticket-system/internal/models"

	"gorm.io/gorm"
)

// AuthHandler holds the dependencies needed by the authentication endpoints.
// Instead of using global variables, we pass dependencies through this struct,
// which is a clean pattern known as "dependency injection".
type AuthHandler struct {
	DB        *gorm.DB      // Database connection for reading/writing user records.
	JWTSecret string        // The secret key used to sign JWT tokens.
	JWTExpiry time.Duration // How long tokens remain valid (e.g., 24 hours).
}

// registerRequest defines the expected shape of the JSON body for POST /auth/register.
// The `json` tags tell Go how to map JSON field names to struct fields.
type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginRequest defines the expected shape of the JSON body for POST /auth/login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginResponse defines the JSON structure returned after a successful login.
type loginResponse struct {
	Token string `json:"token"`
}

// Register handles POST /auth/register — creates a new user account.
//
// Expected request body:
//
//	{
//	    "username": "johndoe",
//	    "email": "john@example.com",
//	    "password": "securepassword123"
//	}
//
// Possible responses:
//   - 201 Created:      User was successfully registered.
//   - 400 Bad Request:  Missing or empty required fields.
//   - 409 Conflict:     Username or email already exists.
//   - 500 Internal:     Unexpected server error.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Step 1: Parse the JSON request body.
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// Step 2: Validate that all required fields are provided.
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Email == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "username, email, and password are required",
		})
		return
	}

	// Step 3: Check if a user with the same username or email already exists.
	var existingUser models.User
	if result := h.DB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser); result.Error == nil {
		// A user was found — this means the username or email is taken.
		respondJSON(w, http.StatusConflict, map[string]string{
			"error": "username or email already exists",
		})
		return
	}

	// Step 4: Hash the password so we never store the plain-text version.
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to hash password",
		})
		return
	}

	// Step 5: Create the new user record in the database.
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if result := h.DB.Create(&user); result.Error != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create user",
		})
		return
	}

	// Step 6: Return the newly created user (password is automatically hidden
	// because of the `json:"-"` tag on the Password field).
	respondJSON(w, http.StatusCreated, user)
}

// Login handles POST /auth/login — authenticates a user and returns a JWT token.
//
// Expected request body:
//
//	{
//	    "email": "john@example.com",
//	    "password": "securepassword123"
//	}
//
// Possible responses:
//   - 200 OK:           Login successful, JWT token returned.
//   - 400 Bad Request:  Missing or empty required fields.
//   - 401 Unauthorized: Email not found or password doesn't match.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Step 1: Parse the JSON request body.
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// Step 2: Validate that required fields are provided.
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "email and password are required",
		})
		return
	}

	// Step 3: Look up the user by email.
	var user models.User
	if result := h.DB.Where("email = ?", req.Email).First(&user); result.Error != nil {
		// User not found — but we don't reveal whether the email exists or not
		// for security reasons. We just say "invalid credentials".
		respondJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "invalid email or password",
		})
		return
	}

	// Step 4: Compare the provided password with the stored hash.
	if !auth.CheckPassword(req.Password, user.Password) {
		respondJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "invalid email or password",
		})
		return
	}

	// Step 5: Generate a JWT token for the authenticated user.
	token, err := auth.GenerateToken(user.ID, h.JWTSecret, h.JWTExpiry)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to generate token",
		})
		return
	}

	// Step 6: Return the token to the client.
	respondJSON(w, http.StatusOK, loginResponse{Token: token})
}

// respondJSON is a helper function that sends a JSON response with the given
// HTTP status code and payload. It's used throughout the handlers to avoid
// repeating the same boilerplate code.
func respondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}
