package handlers

import (
	"db-client/internal/stores"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

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


// Builds a filter object based on query parameters and hands it to the venueStore, which
// will fetch venues accordingly. 
func (h *VenueHandler) List(w http.ResponseWriter, r *http.Request) {
    filter := stores.VenueListFilter{}

	category := r.URL.Query().Get("category")
    if category != "" {
        filter.Category = &category
    }

    beverageNames := r.URL.Query()["names"]  // Returns a slice
    if len(beverageNames) != 0 {
        fmt.Println("DEBUG BEVERAGE PARAMETER:", beverageNames)
        filter.BeverageNames = &beverageNames
    }

    maxPriceString := r.URL.Query().Get("max_price")
    if maxPriceString != "" {
        price, err := strconv.Atoi(maxPriceString)
        if err != nil {
            http.Error(w, "invalid value for max_price", http.StatusBadRequest)
            return
        }
        filter.MaxPrice = &price
    }

    // Time is the only required parameter, the default in the client should be the current time
    timeString := r.URL.Query().Get("time") // hh:mm format, ex 21:30, 09:50
    if timeString == "" {
        http.Error(w, "time parameter missing", http.StatusBadRequest)
        return
    }
    
    t, err := time.Parse("15:04", timeString)
    if err != nil {
        http.Error(w, "invalid time format", http.StatusBadRequest)
        return
    }
    filter.Time = t
    
    onlyHappyHourString := r.URL.Query().Get("happy_hour")
    if onlyHappyHourString != "" {
        val, err := strconv.ParseBool(onlyHappyHourString)
        if err != nil {
            http.Error(w, "invalid value for happy_hour", http.StatusBadRequest)
            return
        }
        filter.OnlyHappyHour = val
    }

    venues, err := h.store.List(r.Context(), filter)
    if err != nil {
        log.Printf("VenueHandler.List: %v", err)
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(venues)
}
