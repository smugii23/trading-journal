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
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("Warning: No .env file found or failed to load it")
	}
	os.Exit(m.Run())
}

func TestAddTrade(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	log.Println("DATABASE_URL inside TestAddTrade:", dbURL) // Print here

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
	id, err := AddTrade(tx, testTrade)
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

func TestListTrades(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect to the test database: %v", err)
	}
	defer db.Close()

	// open a transaction for the test to rollback
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// insert test trade data
	mockTrades := []Trade{
		{Ticker: "AAPL", EntryPrice: 150.0, ExitPrice: 155.0, Quantity: 10, TradeDate: time.Now(), StopLoss: 145.0, TakeProfit: 160.0, Notes: "Test Trade 1", Screenshot: "screenshot1.png"},
		{Ticker: "GOOGL", EntryPrice: 2800.0, ExitPrice: 2900.0, Quantity: 5, TradeDate: time.Now(), StopLoss: 2750.0, TakeProfit: 2950.0, Notes: "Test Trade 2", Screenshot: "screenshot2.png"},
	}

	for _, trade := range mockTrades {
		_, err := AddTrade(tx, trade)
		if err != nil {
			t.Fatalf("failed to insert mock trade: %v", err)
		}
	}

	// retrieve only AAPL trades
	filter := TradeFilter{
		Ticker: "AAPL",
	}
	trades, err := ListTrades(tx, filter)
	if err != nil {
		t.Fatalf("listTrades failed: %v", err)
	}

	// validate result
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].Ticker != "AAPL" {
		t.Errorf("expected ticker AAPL, got %s", trades[0].Ticker)
	}

	// test profit filter
	minProfit := 50.0
	filter = TradeFilter{
		MinProfit: &minProfit,
	}
	trades, err = ListTrades(tx, filter)
	if err != nil {
		t.Fatalf("listTrades failed: %v", err)
	}

	if len(trades) != 2 {
		t.Fatalf("expected 2 trades with min profit 50, got %d", len(trades))
	}
}

func TestGetTrade(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect to the test database: %v", err)
	}
	defer db.Close()

	testTrade := Trade{
		Ticker:     "ES",
		EntryPrice: 5500.5,
		ExitPrice:  5505.5,
		Quantity:   4,
		TradeDate:  time.Now(),
		StopLoss:   5497.5,
		TakeProfit: 5505.5,
		Notes:      "5 point scalp",
		Screenshot: "",
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	id, err := AddTrade(tx, testTrade)
	if err != nil {
		t.Fatalf("failed to add trade: %v", err)
	}
	trade, err := GetTrade(tx, id)
	if err != nil {
		t.Fatalf("failed to get trade: %v", err)
	}
	if trade.Ticker != testTrade.Ticker {
		t.Errorf("expected ticker %s, got %s", testTrade.Ticker, trade.Ticker)
	}
	if trade.EntryPrice != testTrade.EntryPrice {
		t.Errorf("expected entry price %f, got %f", testTrade.EntryPrice, trade.EntryPrice)
	}
	if trade.ExitPrice != testTrade.ExitPrice {
		t.Errorf("expected exit price %f, got %f", testTrade.ExitPrice, trade.ExitPrice)
	}
	if trade.Quantity != testTrade.Quantity {
		t.Errorf("expected quantity %f, got %f", testTrade.Quantity, trade.Quantity)
	}
	if trade.StopLoss != testTrade.StopLoss {
		t.Errorf("expected stop loss %f, got %f", testTrade.StopLoss, trade.StopLoss)
	}
	if trade.TakeProfit != testTrade.TakeProfit {
		t.Errorf("expected take profit %f, got %f", testTrade.TakeProfit, trade.TakeProfit)
	}
	if trade.Notes != testTrade.Notes {
		t.Errorf("expected notes %s, got %s", testTrade.Notes, trade.Notes)
	}
	if trade.Screenshot != testTrade.Screenshot {
		t.Errorf("expected screenshot %s, got %s", testTrade.Screenshot, trade.Screenshot)
	}
}
