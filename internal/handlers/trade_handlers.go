package handlers

import (
	"database/sql"
	"encoding/json"
	"internal/models"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
}

func listTradesHandler(w http.ResponseWriter, r *http.Request) {
	// parse the query parameters
	var filter models.TradeFilter
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	// get the trades
	trades, err := models.ListTrades(db, filter)
	if err != nil {
		http.Error(w, "failed to list trades", http.StatusInternalServerError)
		return
	}
	// return the trades
	if err := json.NewEncoder(w).Encode(trades); err != nil {
		http.Error(w, "failed to encode trades", http.StatusInternalServerError)
		return
	}
}

func addTradesHandler(w http.ResponseWriter, r *http.Request) {}
