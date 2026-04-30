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
