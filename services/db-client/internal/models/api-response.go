package models

import (
	"time"

	"github.com/google/uuid"
)

// Returned when requesting a list of submissions, does not include
// payload, only metadata
type ListSubmissionsResponse struct {
    Submissions []*SubmissionListItem `json:"submissions"`
}
type SubmissionListItem struct {
    ID          uuid.UUID  `json:"id"`
    SubmittedBy uuid.UUID  `json:"submitted_by"`
    Category    string     `json:"category"`
    Status      string     `json:"status"`
    ReviewedAt  *time.Time `json:"reviewed_at"`
    CreatedAt   time.Time  `json:"created_at"`
}



// Returned by /database/venue/list
type FilterVenuesResponse struct {
    Venues []*FilteredVenue `json:"venues"`
}

type FilteredVenue struct {
    ID          string     `json:"id"`  
    VenueName   string     `json:"venue_name"`
    ChainName   *string   `json:"chain_name,omitempty"`
    Location    FilteredVenueLocation   `json:"location"`
    Hours       VenueHours  `json:"hours"`
    MatchedUnit  MatchedUnit `json:"matched_unit"`
}

type VenueHours struct {
    OpenTime        string   `json:"open_time"`
    ClosingTime     string   `json:"closing_time"`
    HasHappyHour    bool        `json:"has_happy_hour"`
    HappyHourStart  string  `json:"happyhour_start,omitempty"`
    HappyHourEnd    string  `json:"happyhour_end,omitempty"`
}

type FilteredVenueLocation struct {
    Street    string    `json:"street"`
    Area      string    `json:"area"`
    City      string    `json:"city"`
    Country   string    `json:"country"`
    Zip       string    `json:"zip"`
    Lat       float64   `json:"lat"`
    Lng       float64   `json:"lng"`
}

type MatchedUnit struct {
    UnitName            string      `json:"unit_name"`
    BeverageName        string      `json:"beverage_name"`
    UnitType            *string     `json:"unit_type"`
    ABV                 *float32    `json:"abv"`
    VolumeML            *int        `json:"volume_ml"`
    Size                *string     `json:"size"`
    Price               float64         `json:"price"`
    BeverageDescription *string     `json:"description"`
}


// Returned when requesting a specific venue by id,
// contains joined location data as well
type GetVenueByIDResponse struct {
    ID          string     `json:"id"`
    Name        string     `json:"name"`
    Location    Location   `json:"location"`
    VenueChainID *string   `json:"venue_chain_id,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}
