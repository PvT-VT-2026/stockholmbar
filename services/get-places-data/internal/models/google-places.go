package models

import "time"


type PlaceSearchRequest struct {
	TextQuery    string `json:"textQuery"`
	LanguageCode string `json:"languageCode"`
}

type PlaceSearchResponse struct {
	Places []struct {
		Id          string `json:"id"`
		DisplayName struct {
			Text string `json:"text"`
		} `json:"displayName"`
		FormattedAddress string `json:"formattedAddress"`
	} `json:"places"`
}

type SearchResultItem struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

type OpeningHours struct {
    DayOfWeek int       `json:"day"`
    OpenTime  time.Time `json:"open"`
    CloseTime time.Time `json:"close"`
}

type PlaceDetailsResponse struct {
	Id          string `json:"id"`
	DisplayName struct {
		Text string `json:"text"`
	} `json:"displayName"`
	Rating   float64 `json:"rating"`
	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
	AddressComponents []struct {
		LongText  string   `json:"longText"`
		ShortText string   `json:"shortText"`
		Types     []string `json:"types"`
	} `json:"addressComponents"`
	RegularOpeningHours struct {
		Periods []struct {
			Open struct {
				Day    int `json:"day"`
				Hour   int `json:"hour"`
				Minute int `json:"minute"`
			} `json:"open"`
			Close struct {
				Day    int `json:"day"`
				Hour   int `json:"hour"`
				Minute int `json:"minute"`
			} `json:"close"`
		} `json:"periods"`
	} `json:"regularOpeningHours"`
}

type PlaceInfo struct {
	PlaceID      string   `json:"place_id"`
	Name         string   `json:"name"`
	Street       string   `json:"street"`
	Area         string   `json:"area"`
	City         string   `json:"city"`
	Country      string   `json:"country"`
	Zip          string   `json:"zip"`
	Lat          float64  `json:"lat"`
	Lng          float64  `json:"lng"`
	Rating       float64  `json:"rating"`
	OpeningHours []OpeningHours `json:"opening_hours"`
}
