package models

import (
	"database/sql"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func TestMain(m *testing.M) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found")
	}
	os.Exit(m.Run())
}

func TestAddTrade(t *testing.T) {
	dbURL := "postgres://postgres:ds89fyphas@localhost:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	testTrade := Trade{
		Ticker:     "NVDA",
		EntryPrice: 150.25,
		ExitPrice:  155.75,
		Quantity:   10,
		TradeDate:  time.Now(),
		StopLoss:   145.50,
		TakeProfit: 155.75,
		Notes:      "Test trade",
		Screenshot: "http://example.com/screenshot.png",
	}

	// test addTrade function
	id, err := addTrade(tx, testTrade)
	if err != nil {
		t.Fatalf("Failed to add trade: %v", err)
	}

	// make sure id is above 0
	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}

	// make sure the trade was added
	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM trades WHERE id = $1", id).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query trade: %v", err)
	}

	// make sure it only returns one trade
	if count != 1 {
		t.Errorf("Expected 1 trade with ID %d, got %d", id, count)
	}
}
