package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
	"bytes"      
	"os"
	"get-places-data/internal/models"
	"io"
)

func GetBarInfoHandler(w http.ResponseWriter, r *http.Request) {
	key := os.Getenv("GOOGLE_API_KEY")
	if key == "" {
		http.Error(w, "API key not set", http.StatusInternalServerError)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing name", http.StatusBadRequest)
		return
	}

	placeID, err := getPlaceID(name, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	barInfo, err := getBarInfo(placeID, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(barInfo)
}

func getPlaceID(name string, apiKey string) (string, error) {
	client := &http.Client{}
	searchURL := "https://places.googleapis.com/v1/places:searchText"

	reqBody, _ := json.Marshal(models.PlaceSearchRequest{TextQuery: name})
	req, err := http.NewRequest("POST", searchURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", apiKey)
	req.Header.Set("X-Goog-FieldMask", "places.id")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var searchData models.PlaceSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchData); err != nil {
		return "", err
	}

	if len(searchData.Places) == 0 {
		return "", fmt.Errorf("could not find place with name: %s", name)
	}

	return searchData.Places[0].Id, nil
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


