package handlers

import (
	"db-client/internal/clients"
	"encoding/json"
	"log"
	"net/http"
)

type PlacesHandler struct {
	client *clients.PlacesClient
}

func NewPlacesHandler(c *clients.PlacesClient) *PlacesHandler {
	return &PlacesHandler{client: c}
}

func (h *PlacesHandler) Search(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "places service not configured", http.StatusServiceUnavailable)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing name parameter", http.StatusBadRequest)
		return
	}

	results, err := h.client.FindPlace(name)
	if err != nil {
		log.Printf("PlacesHandler.Search: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *PlacesHandler) Info(w http.ResponseWriter, r *http.Request) {
	if h.client == nil {
		http.Error(w, "places service not configured", http.StatusServiceUnavailable)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	info, err := h.client.GetPlaceInfo(id)
	if err != nil {
		log.Printf("PlacesHandler.Info: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}
