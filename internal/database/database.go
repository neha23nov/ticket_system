// Package database handles the connection to the SQLite database and
// ensures the required tables exist before the application starts serving requests.
//
// We use GORM (a popular Go ORM — Object-Relational Mapper) to interact with
// the database. GORM lets us work with Go structs instead of writing raw SQL,
// making the code easier to read and maintain.
//
// SQLite is a file-based database — it stores everything in a single file
// (e.g., "data/tickets.db") with zero configuration needed. This makes it
// perfect for small services and Docker deployments.
package database

import (
	"log"

	"ticket-system/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// InitDB opens (or creates) the SQLite database file at the given path
// and automatically creates or updates the "users" and "tickets" tables
// to match the struct definitions in the models package.
//
// Parameters:
//   - dbPath: The file path for the SQLite database (e.g., "data/tickets.db").
//
// Returns a *gorm.DB instance that handlers use to read/write data.
// If the database cannot be opened, the application will terminate with a fatal error.
func InitDB(dbPath string) *gorm.DB {
	// Open (or create) the SQLite database file.
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// AutoMigrate examines the User and Ticket structs and creates/updates
	// the database tables to match. If the tables already exist, it will
	// add any new columns but won't delete existing ones.
	err = db.AutoMigrate(&models.User{}, &models.Ticket{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database connected and migrated successfully")
	return db
}
