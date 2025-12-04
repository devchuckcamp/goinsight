package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/chuckie/goinsight/internal/config"
	"github.com/chuckie/goinsight/internal/db"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("Running database seeder...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	sqlDB, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer sqlDB.Close()

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// Run seed migration (002_seed_feedback.sql)
	seedFile := "migrations/002_seed_feedback.sql"
	content, err := os.ReadFile(seedFile)
	if err != nil {
		log.Fatalf("Failed to read seed file: %v", err)
	}

	fmt.Println("Executing seed data...")
	_, err = sqlDB.Exec(string(content))
	if err != nil {
		log.Fatalf("Failed to execute seed data: %v", err)
	}

	fmt.Println("Seed data inserted successfully!")

	// Show some stats
	var count int
	err = sqlDB.QueryRow("SELECT COUNT(*) FROM feedback_enriched").Scan(&count)
	if err != nil {
		log.Printf("Warning: Could not get count: %v", err)
	} else {
		fmt.Printf("Total feedback records in database: %d\n", count)
	}

	// Use the db package to avoid unused import error
	_ = db.RunMigrations
}
