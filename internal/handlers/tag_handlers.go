package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"trading-journal/internal/models"

	"github.com/go-chi/chi/v5"
)

type TagHandlers struct {
	db *sql.DB
}

func NewTagHandlers(db *sql.DB) *TagHandlers {
	return &TagHandlers{db: db}
}

// handler to create a new tag
func (h *TagHandlers) CreateTagHandler(w http.ResponseWriter, r *http.Request) {
	var tag models.Tag
	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Received tag creation request: %+v", tag)

	// in the future, i'll implement user auth. for now, i'll just use 1 as userid
	tag.UserID = 1

	if err := models.CreateTag(h.db, &tag); err != nil {
		log.Printf("Error creating tag in database: %v", err)
		http.Error(w, "Failed to create tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(tag); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TagHandlers) ListTagsHandler(w http.ResponseWriter, r *http.Request) {
	userID := 1

	tags, err := models.GetTagsByUserID(h.db, userID)
	if err != nil {
		http.Error(w, "Failed to retrieve tags: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tags); err != nil {
		http.Error(w, "Failed to encode tags", http.StatusInternalServerError)
		return
	}
}
func (h *TagHandlers) GetTagHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	userID := 1

	// get all tags from a user and find the one that we need
	tags, err := models.GetTagsByUserID(h.db, userID)
	if err != nil {
		http.Error(w, "Failed to retrieve tags: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var foundTag *models.Tag
	for _, tag := range tags {
		if tag.ID == id {
			foundTag = &tag
			break
		}
	}

	if foundTag == nil {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(foundTag); err != nil {
		http.Error(w, "Failed to encode tag", http.StatusInternalServerError)
		return
	}
}

func (h *TagHandlers) UpdateTagHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	var tag models.Tag
	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	tag.ID = id
	tag.UserID = 1

	if err := models.UpdateTag(h.db, &tag); err != nil {
		http.Error(w, "Failed to update tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tag); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *TagHandlers) DeleteTagHandler(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}
	userID := 1

	if err := models.DeleteTag(h.db, id, userID); err != nil {
		http.Error(w, "Failed to delete tag: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TagHandlers) AddTagToTradeHandler(w http.ResponseWriter, r *http.Request) {
	tradeIDStr := chi.URLParam(r, "trade_id")
	tradeID, err := strconv.Atoi(tradeIDStr)
	if err != nil {
		http.Error(w, "Invalid trade ID", http.StatusBadRequest)
		return
	}

	tagIDStr := chi.URLParam(r, "tag_id")
	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	if err := models.AddTagToTrade(h.db, tradeID, tagID); err != nil {
		http.Error(w, "Failed to add tag to trade: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TagHandlers) RemoveTagFromTradeHandler(w http.ResponseWriter, r *http.Request) {
	tradeIDStr := chi.URLParam(r, "trade_id")
	tradeID, err := strconv.Atoi(tradeIDStr)
	if err != nil {
		http.Error(w, "Invalid trade ID", http.StatusBadRequest)
		return
	}

	tagIDStr := chi.URLParam(r, "tag_id")
	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		http.Error(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	if err := models.RemoveTagFromTrade(h.db, tradeID, tagID); err != nil {
		http.Error(w, "Failed to remove tag from trade: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TagHandlers) GetTradeTagsHandler(w http.ResponseWriter, r *http.Request) {
	tradeIDStr := chi.URLParam(r, "trade_id")
	tradeID, err := strconv.Atoi(tradeIDStr)
	if err != nil {
		http.Error(w, "Invalid trade ID", http.StatusBadRequest)
		return
	}

	tags, err := models.GetTagsByTradeID(h.db, tradeID)
	if err != nil {
		http.Error(w, "Failed to retrieve tags: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tags); err != nil {
		http.Error(w, "Failed to encode tags", http.StatusInternalServerError)
		return
	}
}
