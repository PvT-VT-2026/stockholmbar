package handlers

import (
	"db-client/internal/models"
	"db-client/internal/stores"
	"encoding/json"
	"fmt"
	"net/http"
)

type UnitHandler struct {
	store *stores.UnitStore
}

func NewUnitHandler(s *stores.UnitStore) *UnitHandler {
	return &UnitHandler{store: s}
}

func (h *UnitHandler) CreateUnits(w http.ResponseWriter, r *http.Request) {
	var input models.CreateUnitsRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println("UnitHandler/CreateUnits: invalid request body: %w", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.store.CreateUnits(r.Context(), input)
	if err != nil {
		fmt.Println("UnitHandler/CreateUnits: %w", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}