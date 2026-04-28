package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"get-places-data/internal/models"
	"os"
)

func GetBarIdsHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	results, err := getBarIds(name, os.Getenv("GOOGLE_API_KEY"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func getBarIds(name string, apiKey string) ([]models.SearchResultItem, error) {
	client := &http.Client{}
	searchURL := "https://places.googleapis.com/v1/places:searchText"

	reqBody, _ := json.Marshal(models.PlaceSearchRequest{TextQuery: name})
	req, err := http.NewRequest("POST", searchURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", apiKey)
	req.Header.Set("X-Goog-FieldMask", "places.id,places.displayName,places.formattedAddress")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var searchData models.PlaceSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchData); err != nil {
		return nil, err
	}

	var results []models.SearchResultItem
	for _, place := range searchData.Places {
		results = append(results, models.SearchResultItem{
			ID:      place.Id,
			Name:    place.DisplayName.Text,
			Address: place.FormattedAddress,
		})
	}

	return results, nil
}