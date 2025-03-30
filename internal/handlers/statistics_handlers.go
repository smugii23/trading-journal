package handlers

import (
	"database/sql"
	"encoding/json"
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

	stats, err := models.GetBasicStats(h.db, userID)
	if err != nil {
		http.Error(w, "Failed to retrieve statistics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, "Failed to encode statistics", http.StatusInternalServerError)
		return
	}
} 