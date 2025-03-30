package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
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
	// parse multipart form data
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		http.Error(w, `{"error": "failed to parse form data"}`, http.StatusBadRequest)
		return
	}

	// extract and convert direction to uppercase
	direction := strings.ToUpper(r.FormValue("direction"))
	if direction != "LONG" && direction != "SHORT" {
		http.Error(w, `{"error": "direction must be LONG or SHORT"}`, http.StatusBadRequest)
		return
	}

	// extract trade data from form
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

	// if the user uploads a screenshot, handle it. in the future, i'm planning to use AWS S3 for this
	file, handler, err := r.FormFile("screenshot")
	if err == nil {
		defer file.Close()

		// create the uploads directory if it doesn't exist
		if err := os.MkdirAll("uploads", 0755); err != nil {
			http.Error(w, `{"error": "failed to create upload directory"}`, http.StatusInternalServerError)
			return
		}

		// create a new file in the uploads directory
		dst, err := os.Create("uploads/" + handler.Filename)
		if err != nil {
			http.Error(w, `{"error": "failed to create destination file"}`, http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// copy the uploaded file to the destination file
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, `{"error": "failed to save uploaded file"}`, http.StatusInternalServerError)
			return
		}

		// set the screenshot url in the trade object
		trade.ScreenshotURL = stringPtr("/uploads/" + handler.Filename)
	} else if err != http.ErrMissingFile {
		http.Error(w, `{"error": "failed to process screenshot"}`, http.StatusBadRequest)
		return
	}

	// add trade to database
	id, err := models.AddTrade(h.db, trade)
	if err != nil {
		log.Printf("Error adding trade: %v", err)
		http.Error(w, `{"error": "failed to add trade"}`, http.StatusInternalServerError)
		return
	}
	trade.ID = id

	// use calculate and insert trademetrics function
	err = models.CalculateAndInsertTradeMetrics(h.db, trade)
	if err != nil {
		log.Printf("Error calculating trade metrics: %v", err)
		http.Error(w, `{"error": "failed to add trade metrics"}`, http.StatusInternalServerError)
		return
	}

	// if successful, return the trade
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(trade)
}

// helper functions
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	// convert the string to a float
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseFloatPtr(s string) *float64 {
	if s == "" {
		return nil
	}
	// convert the string to a float
	f, _ := strconv.ParseFloat(s, 64)
	return &f
}

func parseTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	// convert the string to a time
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
	// get the trade id from the url
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

	// get the trade from the database
	trade, err := models.GetTrade(h.db, idInt)
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

func (h *TradeHandlers) UpdateTradeHandler(w http.ResponseWriter, r *http.Request) {
	// get the trade id from the url
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing trade ID", http.StatusBadRequest)
		return
	}

	// get the trade from the request body
	var trade models.Trade
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// convert the trade id to int
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
	// default values for filter
	var filter models.TradeFilter
	
	// check if request is GET or POST
	if r.Method == "GET" {
		// get limit from query string
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
				return
			}
			filter.Limit = limit
		}
		
		// get offset from query string
		if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
			offset, err := strconv.Atoi(offsetStr)
			if err != nil {
				http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
				return
			}
			filter.Offset = offset
		}
		
		// get ticker from query string
		if ticker := r.URL.Query().Get("ticker"); ticker != "" {
			filter.Ticker = ticker
		}
		
		// get start date from query string
		if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
			startDate, err := time.Parse("2006-01-02", startDateStr)
			if err != nil {
				http.Error(w, "Invalid start_date format (use YYYY-MM-DD)", http.StatusBadRequest)
				return
			}
			filter.StartDate = &startDate
		}
		
		if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
			endDate, err := time.Parse("2006-01-02", endDateStr)
			if err != nil {
				http.Error(w, "Invalid end_date format (use YYYY-MM-DD)", http.StatusBadRequest)
				return
			}
			filter.EndDate = &endDate
		}
		
		// get sort parameters
		if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
			filter.SortBy = sortBy
			if r.URL.Query().Get("sort_desc") == "true" {
				filter.SortDesc = true
			}
		}
	} else if r.Method == "POST" {
		// if POST, get filter from request body
		if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
			http.Error(w, "Failed to decode filter", http.StatusBadRequest)
			return
		}
	}

	// get the trades from the database
	trades, err := models.ListTrades(h.db, filter)
	if err != nil {
		http.Error(w, "Failed to fetch trades: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// return the list of trades
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(trades)
}
