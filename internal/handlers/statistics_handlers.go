package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"trading-journal/internal/models"
)

type StatisticsHandlers struct {
	db *sql.DB
}

func NewStatisticsHandlers(db *sql.DB) *StatisticsHandlers {
	return &StatisticsHandlers{db: db}
}

func (h *StatisticsHandlers) GetStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	// auth will be implemented later, for now i'll use ID 1
	userID := 1

	// enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	stats, err := models.GetBasicStats(h.db, userID)
	if err != nil {
		log.Printf("Error getting statistics for user %d: %+v", userID, err)
		http.Error(w, "Failed to retrieve statistics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Error encoding statistics: %+v", err)
		http.Error(w, "Failed to encode statistics", http.StatusInternalServerError)
		return
	}
} 