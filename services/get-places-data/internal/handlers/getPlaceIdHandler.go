package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"get-places-data/internal/models"
	"io"
	"net/http"
)

func (env *APIEnv) GetPlaceIdsHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	results, err := getPlaceIds(name, env.GoogleAPIKey, "https://places.googleapis.com")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func getPlaceIds(name string, apiKey string, baseUrl string) ([]models.SearchResultItem, error) {
	client := &http.Client{}
	searchURL := fmt.Sprintf("%s/v1/places:searchText", baseUrl)

	reqBody, _ := json.Marshal(models.PlaceSearchRequest{TextQuery: name, LanguageCode: "en"})
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google Places searchText returned %d: %s", resp.StatusCode, body)
	}

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