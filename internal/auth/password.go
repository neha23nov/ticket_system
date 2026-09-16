// Package auth provides helper functions for secure password handling
// and JWT token management used by the authentication system.
//
// This file handles password security using bcrypt, which is an industry-standard
// algorithm for hashing passwords. It ensures that even if the database is
// compromised, the actual passwords remain safe.
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword takes a plain-text password (e.g., "mypassword123") and returns
// a secure, one-way hash that can be safely stored in the database.
//
// The hash looks something like: "$2a$10$N9qo8uLOickgx2ZMRZoMye..."
// It is impossible to reverse this hash back into the original password.
//
// bcrypt.DefaultCost (10) controls how computationally expensive the hashing
// is — higher values are more secure but slower. The default is a good balance.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a plain-text password against a previously hashed
// password to see if they match.
//
// This is used during login: we take the password the user typed in, and
// compare it to the hash we stored when they registered.
//
// Returns true if the password is correct, false otherwise.
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
