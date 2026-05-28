package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Request body for requesting to create a submission.
type CreateSubmissionRequest struct {
    Category  string          `json:"category"`   // "unit", "venue", etc
    Payload   json.RawMessage `json:"payload"`    // the raw submission data
}

type CreateVenuePayload struct {
    Name         string     `json:"name"`
    VenueChainID *uuid.UUID `json:"venue_chain_id,omitempty"`
    Street       string     `json:"street"`
    Area         string     `json:"area"`
    City         string     `json:"city"`
    Country      string     `json:"country"`
    Zip          string     `json:"zip"`
    Lat          float64    `json:"lat"`
    Lng          float64    `json:"lng"`
}

type CreateUnitsPayload struct {
    VenueID *uuid.UUID `json:"venueID"`
    Units []*UnitInput `json:"units"`
    Image    *string    `json:"image,omitempty"`    // raw base64 string from client
    ImageURL *string    `json:"imageUrl,omitempty"` // pre-uploaded Supabase Storage URL
}

type UnitInput struct {
    Name       string     `json:"name"`
    VolumeML   *int       `json:"volume_ml"`
    Size        *string    `json:"size"`
    UnitType   *string    `json:"unit_type"`
    Price      int      `json:"price"`
    Currency   string  `json:"currency"`
    ABV        float32  `json:"abv"`
}

// Partial update for a venue's name and location fields.
// Only non-nil fields are applied.
type UpdateVenueInput struct {
    Name    *string `json:"name"`
    Street  *string `json:"street"`
    Area    *string `json:"area"`
    City    *string `json:"city"`
    Zip     *string `json:"zip"`
    Country *string `json:"country"`
}

// Partial update for a venue_unit (menu item).
// Only non-nil fields are applied.
// Amount triggers a new price_record row (time-series history).
type UpdateMenuItemInput struct {
    BeverageName *string  `json:"beverage_name"`
    UnitName     *string  `json:"unit_name"`
    UnitType     *string  `json:"unit_type"`
    VolumeMl     *int     `json:"volume_ml"`
    ABV          *float64 `json:"abv"`
    Amount       *float64 `json:"amount"`
    Currency     *string  `json:"currency"`
}
