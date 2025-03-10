package models

import (
	"database/sql"
	"testing"
	"time"
	"trading-journal/config"

	_ "github.com/lib/pq"
)

func TestAddTrade(t *testing.T) {
	dbURL := config.GetEnv("DATABASE_URL", "")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create the trades table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS trades (
		id SERIAL PRIMARY KEY,
		ticker TEXT NOT NULL,
		entry_price NUMERIC NOT NULL,
		exit_price NUMERIC,
		quantity NUMERIC NOT NULL,
		trade_date TIMESTAMP NOT NULL,
		stop_loss NUMERIC,
		take_profit NUMERIC,
		notes TEXT,
		screenshot_url TEXT
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Create a sample trade
	testTrade := Trade{
		Ticker:     "AAPL",
		EntryPrice: 150.25,
		ExitPrice:  155.75,
		Quantity:   10,
		TradeDate:  time.Now(),
		StopLoss:   145.50,
		TakeProfit: 160.00,
		Notes:      "Test trade",
		Screenshot: "http://example.com/screenshot.png",
	}

	// Call the function we're testing
	id, err := addTrade(db, testTrade)
	if err != nil {
		t.Fatalf("Failed to add trade: %v", err)
	}

	// Verify the ID is greater than 0
	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}

	// Verify the trade was actually added
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM trades WHERE id = $1", id).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query trade: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 trade with ID %d, got %d", id, count)
	}
}
