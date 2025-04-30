package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"trading-journal/internal/services"
)

func ImportNinjatraderTradesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	importer := services.NewNinjaTraderImporterService()

	importedTrades, err := importer.ImportTrades()
	if err != nil {
		log.Printf("Error importing NinjaTrader trades: %v", err)
		http.Error(w, fmt.Sprintf("Failed to import trades: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// TODO: make it save these imported trades to the database after I fix these bugs

	response := map[string]interface{}{
		"message":       "Trade import process completed.",
		"importedCount": len(importedTrades),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding import response: %v", err)
	}
	log.Printf("Successfully handled NinjaTrader import request. Found %d potential trades.", len(importedTrades))
}
