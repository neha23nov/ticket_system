// Package models defines the data structures (database tables) used throughout
// the ticket system application.
//
// This file defines the Ticket model — each ticket represents a task or issue
// created by a user. Tickets go through a simple lifecycle:
//
//	open  →  in_progress  →  closed
//
// Once a ticket is closed, it cannot be reopened.
package models

import "time"

// These constants define the three allowed statuses for a ticket.
// Using constants prevents typos and makes the code easier to maintain.
const (
	StatusOpen       = "open"        // The ticket has been created but work hasn't started.
	StatusInProgress = "in_progress" // Someone is actively working on this ticket.
	StatusClosed     = "closed"      // The ticket is resolved and cannot be reopened.
)

// Ticket represents a support/task ticket in the system.
//
// Fields:
//   - ID:          A unique number automatically assigned to each ticket.
//   - Title:       A short summary of what the ticket is about.
//   - Description: A detailed explanation of the issue or task.
//   - Status:      The current state of the ticket (open, in_progress, or closed).
//                  Defaults to "open" when a new ticket is created.
//   - UserID:      The ID of the user who created this ticket. This is used to
//                  enforce ownership — only the creator can view or update it.
//   - CreatedAt:   When the ticket was first created.
//   - UpdatedAt:   When the ticket was last modified (e.g., status change).
type Ticket struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"not null"`
	Description string    `json:"description" gorm:"not null"`
	Status      string    `json:"status" gorm:"default:open;not null"`
	UserID      uint      `json:"user_id" gorm:"not null;index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ValidStatuses is a quick-lookup map to check whether a given string is a
// valid ticket status. For example: ValidStatuses["open"] returns true.
var ValidStatuses = map[string]bool{
	StatusOpen:       true,
	StatusInProgress: true,
	StatusClosed:     true,
}

// AllowedTransitions defines which status changes are permitted.
// The key is the current status, and the value is the only status it can move to.
//
// Rules:
//   - "open"        can only move to "in_progress"
//   - "in_progress" can only move to "closed"
//   - "closed"      cannot move to anything (not in this map at all)
var AllowedTransitions = map[string]string{
	StatusOpen:       StatusInProgress,
	StatusInProgress: StatusClosed,
}

// IsValidTransition checks whether moving from currentStatus to newStatus
// is allowed by the business rules defined above.
//
// Returns true only if the transition is permitted, false otherwise.
func IsValidTransition(currentStatus, newStatus string) bool {
	allowedNext, exists := AllowedTransitions[currentStatus]
	if !exists {
		// Current status has no allowed transitions (e.g., "closed")
		return false
	}
	return allowedNext == newStatus
}
