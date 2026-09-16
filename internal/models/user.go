// Package models defines the data structures (database tables) used throughout
// the ticket system application.
//
// This file defines the User model — every person who registers in the system
// gets a record in the "users" table. The User stores their username, email,
// and a securely hashed version of their password (the raw password is never saved).
package models

import "time"

// User represents a registered user in the system.
//
// Fields:
//   - ID:        A unique number automatically assigned to each user.
//   - Username:  The user's chosen display name (must be unique across all users).
//   - Email:     The user's email address (must be unique, used for login).
//   - Password:  The bcrypt-hashed password. The `json:"-"` tag ensures this
//                field is NEVER included when sending user data back in API responses.
//   - CreatedAt: The date and time when the user registered.
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}
