package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"trading-journal/internal/models"

	"github.com/go-chi/chi/v5"
)

type TradeHandlers struct {
	db *sql.DB
}

func NewTradeHandlers(db *sql.DB) *TradeHandlers {
	return &TradeHandlers{db: db}
}

func (h *TradeHandlers) AddTradeHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form data
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		http.Error(w, `{"error": "failed to parse form data"}`, http.StatusBadRequest)
		return
	}

	// Extract and convert direction to uppercase
	direction := strings.ToUpper(r.FormValue("direction"))
	if direction != "LONG" && direction != "SHORT" {
		http.Error(w, `{"error": "direction must be LONG or SHORT"}`, http.StatusBadRequest)
		return
	}

	// Extract trade data from form
	trade := models.Trade{
		Ticker:       r.FormValue("ticker"),
		Direction:    direction,
		EntryPrice:   parseFloat(r.FormValue("entry_price")),
		ExitPrice:    parseFloat(r.FormValue("exit_price")),
		Quantity:     parseFloat(r.FormValue("quantity")),
		TradeDate:    parseTime(r.FormValue("trade_date")),
		EntryTime:    parseTime(r.FormValue("entry_time")),
		ExitTime:     parseTime(r.FormValue("exit_time")),
		StopLoss:     parseFloatPtr(r.FormValue("stop_loss")),
		TakeProfit:   parseFloatPtr(r.FormValue("take_profit")),
		Commissions:  parseFloatPtr(r.FormValue("commissions")),
		HighestPrice: parseFloatPtr(r.FormValue("highest_price")),
		LowestPrice:  parseFloatPtr(r.FormValue("lowest_price")),
		Notes:        stringPtr(r.FormValue("notes")),
	}

	// Handle file upload
	file, handler, err := r.FormFile("screenshot")
	if err == nil {
		defer file.Close()
		// Save the file and set the screenshot URL in the trade object
		trade.ScreenshotURL = stringPtr("/uploads/" + handler.Filename)
	} else if err != http.ErrMissingFile {
		http.Error(w, `{"error": "failed to process screenshot"}`, http.StatusBadRequest)
		return
	}

	// Add trade to database
	id, err := models.AddTrade(h.db, trade)
	if err != nil {
		log.Printf("Error adding trade: %v", err)
		http.Error(w, `{"error": "failed to add trade"}`, http.StatusInternalServerError)
		return
	}
	trade.ID = id

	// Calculate and insert trade metrics
	err = models.CalculateAndInsertTradeMetrics(h.db, trade)
	if err != nil {
		log.Printf("Error calculating trade metrics: %v", err)
		http.Error(w, `{"error": "failed to add trade metrics"}`, http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(trade)
}

// Helper functions
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseFloatPtr(s string) *float64 {
	if s == "" {
		return nil
	}
	f, _ := strconv.ParseFloat(s, 64)
	return &f
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
func (h *TradeHandlers) GetTradeHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing trade ID", http.StatusBadRequest)
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "invalid trade ID", http.StatusBadRequest)
		return
	}

	trade, err := models.GetTrade(h.db, idInt)
	if err != nil {
		http.Error(w, "failed to get trade", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(trade); err != nil {
		http.Error(w, "failed to encode trade", http.StatusInternalServerError)
		return
	}
}

func (h *TradeHandlers) UpdateTradeHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing trade ID", http.StatusBadRequest)
		return
	}

	var trade models.Trade
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "invalid trade ID", http.StatusBadRequest)
		return
	}

	trade.ID = idInt
	if err := models.UpdateTrade(h.db, trade); err != nil {
		http.Error(w, "failed to update trade", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(trade); err != nil {
		http.Error(w, "failed to encode trade", http.StatusInternalServerError)
		return
	}
}

func (h *TradeHandlers) DeleteTradeHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing trade ID", http.StatusBadRequest)
		return
	}
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "invalid trade ID", http.StatusBadRequest)
		return
	}

	if err := models.DeleteTrade(h.db, idInt); err != nil {
		http.Error(w, "failed to delete trade", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TradeHandlers) ListTradesHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filtering
	var filter models.TradeFilter
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		http.Error(w, "failed to decode trades", http.StatusInternalServerError)
		return
	}

	// Fetch trades from the database
	trades, err := models.ListTrades(h.db, filter)
	if err != nil {
		http.Error(w, "failed to fetch trades", http.StatusInternalServerError)
		return
	}

	// Return the list of trades
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(trades)
}
