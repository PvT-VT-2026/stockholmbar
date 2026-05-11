package models

type PlaceSearchRequest struct {
	TextQuery string `json:"textQuery"`
}

type PlaceSearchResponse struct {
	Places []struct {
		Id string `json:"id"`
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

type PlaceDetailsResponse struct {
	Id string `json:"id"`
	DisplayName struct {
		Text string `json:"text"`
	} `json:"displayName"`
	Rating float64 `json:"rating"`
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
		WeekdayDescriptions []string `json:"weekdayDescriptions"`
	} `json:"regularOpeningHours"`
}

type PlaceInfo struct {
	PlaceID      string  `json:"place_id"`
	Name         string  `json:"name"`
	Street       string  `json:"street"`
	Area         string  `json:"area"`
	City         string  `json:"city"`
	Country      string  `json:"country"`
	Zip          string  `json:"zip"`
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
	Rating       float64  `json:"rating"`       
	OpeningHours []string `json:"opening_hours"`
}

