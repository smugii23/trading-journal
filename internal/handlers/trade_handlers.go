package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"trading-journal/internal/models"

	"github.com/go-chi/chi/v5"
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

func (h *TradeHandlers) AddTradeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var trade models.Trade
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	id, err := models.AddTrade(h.db, trade)
	if err != nil {
		http.Error(w, "failed to add trade", http.StatusInternalServerError)
		return
	}
	trade.ID = id

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/trades/%d", trade.ID))
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(trade); err != nil {
		http.Error(w, "failed to encode trade", http.StatusInternalServerError)
		return
	}
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
