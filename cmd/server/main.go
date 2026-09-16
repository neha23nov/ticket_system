
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"ticket-system/internal/database"
	"ticket-system/internal/handlers"
	"ticket-system/internal/middleware"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	

	port := getEnv("PORT", "8080")                              // Port the server listens on.
	jwtSecret := getEnv("JWT_SECRET", "default-secret-key-123") // Secret key for signing JWT tokens.
	dbPath := getEnv("DB_PATH", "data/tickets.db")              // Path to the SQLite database file.


	jwtExpiry := 24 * time.Hour


	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	
	db := database.InitDB(dbPath)



	authHandler := &handlers.AuthHandler{
		DB:        db,
		JWTSecret: jwtSecret,
		JWTExpiry: jwtExpiry,
	}

	ticketHandler := &handlers.TicketHandler{
		DB: db,
	}


	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	
	r.Get("/health", handlers.HealthCheck)

	
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)


	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth(jwtSecret))

		// Ticket management routes:
		r.Post("/tickets", ticketHandler.Create)              // Create a new ticket
		r.Get("/tickets", ticketHandler.List)                  // List all my tickets
		r.Get("/tickets/{id}", ticketHandler.GetByID)          // Get a specific ticket by ID
		r.Patch("/tickets/{id}/status", ticketHandler.UpdateStatus) // Update a ticket's status
	})

	// -------------------------------------------------------------------------
	// Frontend Static Files
	// -------------------------------------------------------------------------

	// Serve CSS and JavaScript files from the "web" directory at /static/.
	// For example, /static/style.css serves the file web/style.css.
	fileServer := http.FileServer(http.Dir("web"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Serve the main HTML page at the root URL (http://localhost:8080/).
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/index.html")
	})

	// =========================================================================
	// Start Server
	// =========================================================================

	log.Printf("🚀 Ticket System server starting on port %s", port)
	log.Printf("📋 Health check: http://localhost:%s/health", port)

	// Start listening for HTTP requests. This call blocks until the server
	// is stopped (e.g., via Ctrl+C or Docker stop).
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
