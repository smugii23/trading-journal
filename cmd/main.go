package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"trading-journal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	config.LoadEnv()
	databaseURL := config.GetEnv("DATABASE_URL", "")
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v\n", err)
	}
	defer db.Close()

	// run the migrations
	m, err := migrate.New(
		"file://./migrations", // the path to migrations
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migration object: %v\n", err)
	}

	// Check if we need to force a specific version (to handle dirty state)
	if len(os.Args) > 1 && os.Args[1] == "force" {
		version := 0 // Reset to before all migrations (or use 1 to keep the first migration)
		if err := m.Force(version); err != nil {
			log.Fatalf("Failed to force version: %v\n", err)
		}
		fmt.Printf("Migration state forced to version %d\n", version)
		return
	}

	// apply the migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v\n", err)
	}

	fmt.Println("Migrations applied successfully")
}
