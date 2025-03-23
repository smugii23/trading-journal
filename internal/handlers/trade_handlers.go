package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"trading-journal/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TradeHandlers struct {
	db *sql.DB
}

func NewTradeHandlers(db *sql.DB) *TradeHandlers {
	return &TradeHandlers{db: db}
}

func (h *TradeHandlers) ListTradesHandler(w http.ResponseWriter, r *http.Request) {
	var filter models.TradeFilter
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	trades, err := models.ListTrades(h.db, filter)
	if err != nil {
		http.Error(w, "failed to list trades", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(trades); err != nil {
		http.Error(w, "failed to encode trades", http.StatusInternalServerError)
		return
	}
}

const (
	maxUploadSize = 10 * 1024 * 1024 // 10MB
	uploadPath    = "./uploads/screenshots"
)

func init() {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}
}

func (h *TradeHandlers) AddTradeHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create a new trade instance
	var trade models.Trade

	// Parse form fields into trade struct
	trade.Ticker = r.FormValue("ticker")
	trade.Direction = r.FormValue("direction")

	// Parse numeric fields
	if entryPrice, err := strconv.ParseFloat(r.FormValue("entry_price"), 64); err == nil {
		trade.EntryPrice = entryPrice
	}

	if exitPrice, err := strconv.ParseFloat(r.FormValue("exit_price"), 64); err == nil && r.FormValue("exit_price") != "" {
		trade.ExitPrice = exitPrice
	}

	if quantity, err := strconv.ParseFloat(r.FormValue("quantity"), 64); err == nil {
		trade.Quantity = quantity
	}

	// Parse date/time fields
	if tradeDate, err := time.Parse("2006-01-02", r.FormValue("trade_date")); err == nil {
		trade.TradeDate = tradeDate
	}

	if entryTime, err := time.Parse("15:04", r.FormValue("entry_time")); err == nil {
		// Combine date and time
		trade.EntryTime = combineDateAndTime(trade.TradeDate, entryTime)
	}

	if exitTime, err := time.Parse("15:04", r.FormValue("exit_time")); err == nil && r.FormValue("exit_time") != "" {
		trade.ExitTime = combineDateAndTime(trade.TradeDate, exitTime)
	}

	// Parse optional fields
	if r.FormValue("stop_loss") != "" {
		if stopLoss, err := strconv.ParseFloat(r.FormValue("stop_loss"), 64); err == nil {
			trade.StopLoss = &stopLoss
		}
	}

	if r.FormValue("take_profit") != "" {
		if takeProfit, err := strconv.ParseFloat(r.FormValue("take_profit"), 64); err == nil {
			trade.TakeProfit = &takeProfit
		}
	}

	if r.FormValue("commissions") != "" {
		if commissions, err := strconv.ParseFloat(r.FormValue("commissions"), 64); err == nil {
			trade.Commissions = &commissions
		}
	}

	if r.FormValue("highest_price") != "" {
		if highestPrice, err := strconv.ParseFloat(r.FormValue("highest_price"), 64); err == nil {
			trade.HighestPrice = &highestPrice
		}
	}

	if r.FormValue("lowest_price") != "" {
		if lowestPrice, err := strconv.ParseFloat(r.FormValue("lowest_price"), 64); err == nil {
			trade.LowestPrice = &lowestPrice
		}
	}

	notes := r.FormValue("notes")
	trade.Notes = &notes

	// Handle screenshot upload
	file, header, err := r.FormFile("screenshot")
	if err == nil && file != nil {
		defer file.Close()

		// Generate unique filename
		ext := filepath.Ext(header.Filename)
		uniqueID := uuid.New().String()
		filename := uniqueID + ext
		filePath := filepath.Join(uploadPath, filename)

		// Create destination file
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to create file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy file contents
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Save relative path to database
		screenshotPath := "/uploads/screenshots/" + filename
		trade.ScreenshotURL = &screenshotPath
	}

	// Insert trade into database
	id, err := models.AddTrade(h.db, trade)
	if err != nil {
		http.Error(w, "Failed to add trade: "+err.Error(), http.StatusInternalServerError)
		return
	}
	trade.ID = id

	// Handle tag associations
	if tagJSON := r.FormValue("tags"); tagJSON != "" {
		var tagIDs []int
		if err := json.Unmarshal([]byte(tagJSON), &tagIDs); err != nil {
			log.Printf("Failed to parse tag IDs: %v", err)
		} else {
			for _, tagID := range tagIDs {
				if err := models.AddTagToTrade(h.db, trade.ID, tagID); err != nil {
					// Log error but continue with other tags
					log.Printf("Failed to add tag %d to trade %d: %v", tagID, trade.ID, err)
				}
			}
		}
	}

	// Calculate and insert metrics
	if err := models.CalculateAndInsertTradeMetrics(h.db, trade); err != nil {
		http.Error(w, "Failed to add trade metrics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get tags for this trade to include in the response
	tags, err := models.GetTagsByTradeID(h.db, trade.ID)
	if err != nil {
		log.Printf("Failed to get tags for trade: %v", err)
	}

	// Create response structure with trade and tags
	response := struct {
		Trade models.Trade `json:"trade"`
		Tags  []models.Tag `json:"tags,omitempty"`
	}{
		Trade: trade,
		Tags:  tags,
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/trades/%d", trade.ID))
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// Helper function to combine date and time
func combineDateAndTime(date time.Time, timeValue time.Time) time.Time {
	return time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		timeValue.Hour(),
		timeValue.Minute(),
		0, 0,
		time.UTC,
	)
}

// GetTradeScreenshot retrieves the screenshot for a trade
func (h *TradeHandlers) GetTradeScreenshot(w http.ResponseWriter, r *http.Request) {
	tradeIDStr := chi.URLParam(r, "id")
	tradeID, err := strconv.Atoi(tradeIDStr)
	if err != nil {
		http.Error(w, "Invalid trade ID", http.StatusBadRequest)
		return
	}

	// Query the database to get the screenshot URL
	var screenshotURL string
	err = h.db.QueryRow("SELECT screenshot FROM trades WHERE id = $1", tradeID).Scan(&screenshotURL)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Trade not found", http.StatusNotFound)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	if screenshotURL == "" {
		http.Error(w, "No screenshot available for this trade", http.StatusNotFound)
		return
	}

	// Redirect to the actual file
	http.Redirect(w, r, screenshotURL, http.StatusFound)
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

	// Get tags for this trade
	tags, err := models.GetTagsByTradeID(h.db, idInt)
	if err != nil {
		log.Printf("Failed to get tags for trade: %v", err)
	}

	// Return response with trade and tags
	response := struct {
		Trade models.Trade `json:"trade"`
		Tags  []models.Tag `json:"tags,omitempty"`
	}{
		Trade: trade,
		Tags:  tags,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
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
