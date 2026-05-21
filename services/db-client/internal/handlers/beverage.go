package handlers

import (
	"db-client/internal/stores"
	"encoding/json"
	"log"
	"net/http"
)

type BeverageHandler struct {
	store *stores.BeverageStore
}

func NewBeverageHandler(s *stores.BeverageStore) *BeverageHandler {
	return &BeverageHandler{store: s}
}


func (h *BeverageHandler) List(w http.ResponseWriter, r *http.Request) {
	beverages, err := h.store.List(r.Context())
	if err != nil {
		log.Printf("BeverageHandler.List: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(beverages)
}
