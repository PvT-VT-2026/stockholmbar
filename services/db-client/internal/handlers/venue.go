package handlers

import (
	"db-client/internal/models"
	"db-client/internal/stores"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
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

    id := chi.URLParam(r, "id")
    venueID, err := uuid.Parse(id)
    if err != nil {
		log.Printf("VenueHandler.GetByID: %v", err)
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }

    venue, err := h.store.GetByID(r. Context(), venueID)
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


func (h *VenueHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	venueID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("VenueHandler.GetMenu: %v", err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	venue, err := h.store.GetByID(r.Context(), venueID)
	if err != nil {
		log.Printf("VenueHandler.GetMenu: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if venue == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	menu, err := h.store.GetMenuByVenueID(r.Context(), venueID)
	if err != nil {
		log.Printf("VenueHandler.GetMenu: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := models.VenueMenuResponse{
		ID:       venue.ID,
		Name:     venue.Name,
		Location: venue.Location,
		Menu:     menu,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// SearchAll returns every non-deleted venue. An optional ?q= query filters by name.
func (h *VenueHandler) SearchAll(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	items, err := h.store.SearchAll(r.Context(), q)
	if err != nil {
		log.Printf("VenueHandler.SearchAll: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.VenueSearchResponse{Venues: items})
}

// Update applies a partial update to a venue's name and/or location.
func (h *VenueHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	venueID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var input models.UpdateVenueInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.Update(r.Context(), venueID, input); err != nil {
		if errors.Is(err, stores.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		log.Printf("VenueHandler.Update: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete soft-deletes a venue.
func (h *VenueHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	venueID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.store.Delete(r.Context(), venueID); err != nil {
		if errors.Is(err, stores.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		log.Printf("VenueHandler.Delete: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateMenuItem applies a partial update to a menu item (venue_unit).
func (h *VenueHandler) UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	venueID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid venue id", http.StatusBadRequest)
		return
	}
	venueUnitID, err := uuid.Parse(chi.URLParam(r, "unitId"))
	if err != nil {
		http.Error(w, "invalid unit id", http.StatusBadRequest)
		return
	}

	var input models.UpdateMenuItemInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateMenuItem(r.Context(), venueID, venueUnitID, input); err != nil {
		if errors.Is(err, stores.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		log.Printf("VenueHandler.UpdateMenuItem: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteMenuItem soft-deletes a single venue_unit (menu item).
func (h *VenueHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	venueID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid venue id", http.StatusBadRequest)
		return
	}
	venueUnitID, err := uuid.Parse(chi.URLParam(r, "unitId"))
	if err != nil {
		http.Error(w, "invalid unit id", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteMenuItem(r.Context(), venueID, venueUnitID); err != nil {
		if errors.Is(err, stores.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		log.Printf("VenueHandler.DeleteMenuItem: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

    // time is optional; omitting it returns all venues scheduled for today
    timeString := r.URL.Query().Get("time") // hh:mm format, ex 21:30, 09:50
    if timeString != "" {
        t, err := time.Parse("15:04", timeString)
        if err != nil {
            http.Error(w, "invalid time format", http.StatusBadRequest)
            return
        }
        filter.Time = &t
    }
    
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
