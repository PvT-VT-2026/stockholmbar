package handlers

import (
	"encoding/json"
	"fmt"
	"get-places-data/internal/models"
	"io"
	"net/http"
)

func (env *APIEnv) GetPlaceInfoHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing place ID", http.StatusBadRequest)
		return
	}

	placeInfo, err := getPlaceInfo(id, env.GoogleAPIKey, "https://places.googleapis.com")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(placeInfo)
}

func getPlaceInfo(placeID string, apiKey string, baseURL string) (*models.PlaceInfo, error) {
	client := &http.Client{}
	detailsURL := fmt.Sprintf("%s/v1/places/%s", baseURL, placeID)

	req, err := http.NewRequest("GET", detailsURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Goog-Api-Key", apiKey)
	req.Header.Set("X-Goog-FieldMask", "id,displayName,location,addressComponents,rating,regularOpeningHours")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Google Places place details returned %d: %s", resp.StatusCode, body)
	}

	var details models.PlaceDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	return formatPlaceInfo(details, placeID), nil
}

func formatPlaceInfo(details models.PlaceDetailsResponse, placeID string) *models.PlaceInfo {
	var OpeningHoursList []models.OpeningHours // Assuming this is your slice

	for _, p := range details.RegularOpeningHours.Periods {
		OpeningHoursList = append(OpeningHoursList, models.OpeningHours{
			DayOfWeek:   p.Open.Day,
			OpenHour:    p.Open.Hour,
			OpenMinute:  p.Open.Minute,
			CloseHour:   p.Close.Hour,
			CloseMinute: p.Close.Minute,
		})
	}

	placeInfo := &models.PlaceInfo{
		PlaceID:      placeID,
		Name:         details.DisplayName.Text,
		Lat:          details.Location.Latitude,
		Lng:          details.Location.Longitude,
		Rating:       details.Rating,
		OpeningHours: OpeningHoursList,
	}

	for _, comp := range details.AddressComponents {
		for _, t := range comp.Types {
			switch t {
			case "route":
				if placeInfo.Street == "" {
                	placeInfo.Street = comp.LongText
				} else {
                	placeInfo.Street = comp.LongText + " " + placeInfo.Street
				}	
			case "street_number":
				placeInfo.Street += " " + comp.LongText
			case "postal_town":
				placeInfo.City = comp.LongText
			case "sublocality_level_1":
				placeInfo.Area = comp.LongText
			case "postal_code":
				placeInfo.Zip = comp.LongText
			case "country":
				placeInfo.Country = comp.LongText
			}
		}
	}

	return placeInfo
}


