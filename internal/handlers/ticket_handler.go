// Package handlers contains all the HTTP request handlers for the ticket system.
//
// This file handles all ticket-related operations:
//   - Creating a new ticket (POST /tickets)
//   - Listing all tickets for the logged-in user (GET /tickets)
//   - Getting a specific ticket by ID (GET /tickets/{id})
//   - Updating a ticket's status (PATCH /tickets/{id}/status)
//
// IMPORTANT ownership rule: A user can ONLY view and modify tickets they created.
// If user A tries to access a ticket created by user B, they will get a 403 Forbidden.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ticket-system/internal/middleware"
	"ticket-system/internal/models"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// TicketHandler holds the dependencies needed by the ticket endpoints.
type TicketHandler struct {
	DB *gorm.DB // Database connection for reading/writing ticket records.
}

// createTicketRequest defines the expected JSON body for creating a new ticket.
type createTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// updateStatusRequest defines the expected JSON body for updating a ticket's status.
type updateStatusRequest struct {
	Status string `json:"status"`
}

// Create handles POST /tickets — creates a new ticket for the logged-in user.
//
// The ticket is automatically assigned to the user who is making the request
// (identified by their JWT token), and its initial status is set to "open".
//
// Expected request body:
//
//	{
//	    "title": "Fix login bug",
//	    "description": "Users cannot log in with special characters in password"
//	}
//
// Possible responses:
//   - 201 Created:      Ticket was successfully created.
//   - 400 Bad Request:  Missing or empty required fields.
//   - 401 Unauthorized: No valid JWT token provided (handled by middleware).
//   - 500 Internal:     Unexpected server error.
func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Step 1: Get the authenticated user's ID from the request context.
	// This was set by the JWT middleware after validating the token.
	userID, ok := middleware.GetUserID(r)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "user not authenticated",
		})
		return
	}

	// Step 2: Parse the JSON request body.
	var req createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	// Step 3: Validate required fields.
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)

	if req.Title == "" || req.Description == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "title and description are required",
		})
		return
	}

	// Step 4: Create the ticket with status "open" and assign it to the current user.
	ticket := models.Ticket{
		Title:       req.Title,
		Description: req.Description,
		Status:      models.StatusOpen,
		UserID:      userID,
	}

	if result := h.DB.Create(&ticket); result.Error != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create ticket",
		})
		return
	}

	// Step 5: Return the newly created ticket.
	respondJSON(w, http.StatusCreated, ticket)
}

// List handles GET /tickets — returns all tickets belonging to the logged-in user.
//
// This endpoint only returns tickets created by the authenticated user.
// It does NOT return tickets created by other users.
//
// Possible responses:
//   - 200 OK:           Array of the user's tickets (may be empty).
//   - 401 Unauthorized: No valid JWT token provided.
func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	// Step 1: Get the authenticated user's ID.
	userID, ok := middleware.GetUserID(r)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "user not authenticated",
		})
		return
	}

	// Step 2: Fetch all tickets where user_id matches the logged-in user.
	var tickets []models.Ticket
	if result := h.DB.Where("user_id = ?", userID).Find(&tickets); result.Error != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to fetch tickets",
		})
		return
	}

	// Step 3: Return the list (empty array if user has no tickets).
	respondJSON(w, http.StatusOK, tickets)
}

// GetByID handles GET /tickets/{id} — returns a specific ticket by its ID.
//
// The ticket must belong to the logged-in user. If the ticket exists but
// belongs to another user, a 403 Forbidden response is returned.
//
// Possible responses:
//   - 200 OK:           The ticket was found and belongs to the user.
//   - 401 Unauthorized: No valid JWT token provided.
//   - 403 Forbidden:    The ticket belongs to another user.
//   - 404 Not Found:    No ticket with the given ID exists.
func (h *TicketHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// Step 1: Get the authenticated user's ID.
	userID, ok := middleware.GetUserID(r)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "user not authenticated",
		})
		return
	}

	// Step 2: Parse the ticket ID from the URL path (e.g., /tickets/42 → id=42).
	ticketID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid ticket ID",
		})
		return
	}

	// Step 3: Look up the ticket in the database.
	var ticket models.Ticket
	if result := h.DB.First(&ticket, ticketID); result.Error != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{
			"error": "ticket not found",
		})
		return
	}

	// Step 4: Check ownership — does this ticket belong to the logged-in user?
	if ticket.UserID != userID {
		respondJSON(w, http.StatusForbidden, map[string]string{
			"error": "you do not have permission to access this ticket",
		})
		return
	}

	// Step 5: Return the ticket.
	respondJSON(w, http.StatusOK, ticket)
}

// UpdateStatus handles PATCH /tickets/{id}/status — updates a ticket's status.
//
// Status transitions follow strict rules:
//   - "open"        → "in_progress"  ✅ Allowed
//   - "in_progress" → "closed"       ✅ Allowed
//   - "closed"      → anything       ❌ Not allowed (closed tickets stay closed)
//   - Any backward transition         ❌ Not allowed
//
// Expected request body:
//
//	{
//	    "status": "in_progress"
//	}
//
// Possible responses:
//   - 200 OK:           Status was successfully updated.
//   - 400 Bad Request:  Invalid status value or transition not allowed.
//   - 401 Unauthorized: No valid JWT token provided.
//   - 403 Forbidden:    The ticket belongs to another user.
//   - 404 Not Found:    No ticket with the given ID exists.
func (h *TicketHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	// Step 1: Get the authenticated user's ID.
	userID, ok := middleware.GetUserID(r)
	if !ok {
		respondJSON(w, http.StatusUnauthorized, map[string]string{
			"error": "user not authenticated",
		})
		return
	}

	// Step 2: Parse the ticket ID from the URL.
	ticketID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid ticket ID",
		})
		return
	}

	// Step 3: Parse the request body to get the new status.
	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	req.Status = strings.TrimSpace(req.Status)

	// Step 4: Validate that the provided status is one of the allowed values.
	if !models.ValidStatuses[req.Status] {
		respondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid status, must be one of: open, in_progress, closed",
		})
		return
	}

	// Step 5: Look up the ticket in the database.
	var ticket models.Ticket
	if result := h.DB.First(&ticket, ticketID); result.Error != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{
			"error": "ticket not found",
		})
		return
	}

	// Step 6: Check ownership.
	if ticket.UserID != userID {
		respondJSON(w, http.StatusForbidden, map[string]string{
			"error": "you do not have permission to update this ticket",
		})
		return
	}

	// Step 7: Validate the status transition using our business rules.
	if !models.IsValidTransition(ticket.Status, req.Status) {
		// Build a helpful error message depending on the situation.
		if ticket.Status == models.StatusClosed {
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error": "closed tickets cannot be reopened",
			})
		} else {
			respondJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid status transition from " + ticket.Status + " to " + req.Status,
			})
		}
		return
	}

	// Step 8: Update the ticket's status in the database.
	ticket.Status = req.Status
	if result := h.DB.Save(&ticket); result.Error != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to update ticket status",
		})
		return
	}

	// Step 9: Return the updated ticket.
	respondJSON(w, http.StatusOK, ticket)
}
