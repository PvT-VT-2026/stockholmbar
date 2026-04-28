package models

import (
	"time"

	"github.com/google/uuid"
)

type CreateVenueRequest struct {
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

type VenueResponse struct {
    ID          string     `json:"id"`
    Name        string     `json:"name"`
    Location    Location   `json:"location"`
    VenueChainID *string   `json:"venue_chain_id,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateUnitsRequest struct {
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
    ABV        int  `json:"abv"`
}

