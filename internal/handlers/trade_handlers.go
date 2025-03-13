package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
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

func addTradesHandler(w http.ResponseWriter, r *http.Request) {
	// check if POST
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// parse the request body
	var trade models.Trade
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// add the trade
	if err := trade.AddTrade(db); err != nil {
		http.Error(w, "failed to add trade", http.StatusInternalServerError)
		return
	}

	// set the headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/trades/%d", trade.ID))
	w.WriteHeader(http.StatusCreated)

	// return the trade
	if err := json.NewEncoder(w).Encode(trade); err != nil {
		http.Error(w, "failed to encode trade", http.StatusInternalServerError)
		return
	}
}

func getTradeHandler(w http.ResponseWriter, r *http.Request) {
	// get the trade ID
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing trade ID", http.StatusBadRequest)
		return
	}

	// get the trade
	trade, err := models.GetTrade(db, id)
	if err != nil {
		http.Error(w, "failed to get trade", http.StatusInternalServerError)
		return
	}

	// return the trade
	if err := json.NewEncoder(w).Encode(trade); err != nil {
		http.Error(w, "failed to encode trade", http.StatusInternalServerError)
		return
	}
}

func updateTradeHandler(w http.ResponseWriter, r *http.Request) {
	// get the trade ID
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing trade ID", http.StatusBadRequest)
		return
	}

	// parse the request body
	var trade models.Trade
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// update the trade
	if err := trade.UpdateTrade(db, id); err != nil {
		http.Error(w, "failed to update trade", http.StatusInternalServerError)
		return
	}

	// return the trade
	if err := json.NewEncoder(w).Encode(trade); err != nil {
		http.Error(w, "failed to encode trade", http.StatusInternalServerError)
		return
	}
}

func deleteTradeHandler(w http.ResponseWriter, r *http.Request) {
	// get the trade ID
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing trade ID", http.StatusBadRequest)
		return
	}

	// delete the trade
	if err := models.DeleteTrade(db, id); err != nil {
		http.Error(w, "failed to delete trade", http.StatusInternalServerError)
		return
	}

	// return success
	w.WriteHeader(http.StatusNoContent)
}
