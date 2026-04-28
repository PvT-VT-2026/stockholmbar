package handlers

import (
    "encoding/json"
    "fmt"
    "net/http" 
	"get-places-data/internal/models"
	"os"
)

func GetBarInfoHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing place ID", http.StatusBadRequest)
		return
	}

	barRecord, err := getBarInfo(id, os.Getenv("GOOGLE_API_KEY"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(barRecord)
}

func getBarInfo(placeID string, apiKey string) (*models.BarInfo, error) {
	client := &http.Client{}
	detailsURL := fmt.Sprintf("https://places.googleapis.com/v1/places/%s", placeID)

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

	var details models.PlaceDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	return formatBarInfo(details, placeID), nil
}

func formatBarInfo(details models.PlaceDetailsResponse, placeID string) *models.BarInfo {
	barInfo := &models.BarInfo{
		PlaceID:      placeID,
		Name:         details.DisplayName.Text,
		Lat:          details.Location.Latitude,
		Lng:          details.Location.Longitude,
		Rating:       details.Rating,
		OpeningHours: details.RegularOpeningHours.WeekdayDescriptions,
	}

	for _, comp := range details.AddressComponents {
		for _, t := range comp.Types {
			switch t {
			case "route":
				if barInfo.Street == "" {
                	barInfo.Street = comp.LongText
				} else {
                	barInfo.Street = comp.LongText + " " + barInfo.Street
				}	
			case "street_number":
				barInfo.Street += " " + comp.LongText
			case "postal_town":
				barInfo.City = comp.LongText
			case "sublocality_level_1":
				barInfo.Area = comp.LongText
			case "postal_code":
				barInfo.Zip = comp.LongText
			case "country":
				barInfo.Country = comp.LongText
			}
		}
	}

	return barInfo
}


