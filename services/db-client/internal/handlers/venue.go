package handlers

import (
	"db-client/internal/models"
	"db-client/internal/stores"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type VenueHandler struct {
	store *stores.VenueStore
}

func NewVenueHandler(s *stores.VenueStore) *VenueHandler {
	return &VenueHandler{store: s}
}


func (h *VenueHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input models.CreateVenueInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	venue, err := h.store.Create(r.Context(), input)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(venue)
}

func (h *VenueHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id, err := uuid.Parse(r.PathValue("id"))
    if err != nil {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }

    venue, err := h.store.GetByID(r. Context(), id)
    if err != nil {
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }

    if venue == nil {
        http.Error(w, "not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(venue)
}
