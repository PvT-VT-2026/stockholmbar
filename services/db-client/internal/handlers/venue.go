package handlers

import (
	"db-client/internal/stores"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// VenueHandler no longer owns its own insertion endpoint,
// as all insertions go through the submission system.
// This handler will be populated by getters only

type VenueHandler struct {
	store *stores.VenueStore
}

func NewVenueHandler(s *stores.VenueStore) *VenueHandler {
	return &VenueHandler{store: s}
}

func (h *VenueHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id, err := uuid.Parse(r.PathValue("id"))
    if err != nil {
		log.Printf("VenueHandler.GetByID: %v", err)
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }

    venue, err := h.store.GetByID(r. Context(), id)
    if err != nil {
		log.Printf("VenueHandler.GetByID: %v", err)
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
