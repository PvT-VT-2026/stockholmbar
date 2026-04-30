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


// Returned when requesting a specific venue by id,
// contains joined location data as well
type VenueResponse struct {
    ID          string     `json:"id"`
    Name        string     `json:"name"`
    Location    Location   `json:"location"`
    VenueChainID *string   `json:"venue_chain_id,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}


