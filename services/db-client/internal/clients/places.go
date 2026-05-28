package clients

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type PlacesClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewPlacesClient(baseURL string) *PlacesClient {
	return &PlacesClient{baseURL: baseURL, httpClient: &http.Client{}}
}

type PlaceSearchResult struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type OpeningHours struct {
	DayOfWeek   int `json:"day"`
	OpenHour    int `json:"open_hour"`
	OpenMinute  int `json:"open_minute"`
	CloseHour   int `json:"close_hour"`
	CloseMinute int `json:"close_minute"`
}

type PlaceInfo struct {
	PlaceID      string         `json:"place_id"`
	Name         string         `json:"name"`
	Street       string         `json:"street"`
	Area         string         `json:"area"`
	City         string         `json:"city"`
	Country      string         `json:"country"`
	Zip          string         `json:"zip"`
	Lat          float64        `json:"lat"`
	Lng          float64        `json:"lng"`
	Rating       float64        `json:"rating"`
	OpeningHours []OpeningHours `json:"opening_hours"`
}

func (c *PlacesClient) FindPlace(query string) ([]PlaceSearchResult, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/findplace?name=" + url.QueryEscape(query))
	if err != nil {
		return nil, fmt.Errorf("FindPlace: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FindPlace: unexpected status %d", resp.StatusCode)
	}

	var results []PlaceSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("FindPlace: decode: %w", err)
	}
	return results, nil
}

func (c *PlacesClient) GetPlaceInfo(placeID string) (*PlaceInfo, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/placeinfo?id=" + url.QueryEscape(placeID))
	if err != nil {
		return nil, fmt.Errorf("GetPlaceInfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GetPlaceInfo: unexpected status %d", resp.StatusCode)
	}

	var info PlaceInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("GetPlaceInfo: decode: %w", err)
	}
	return &info, nil
}
