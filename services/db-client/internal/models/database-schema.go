package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)


type SubmissionImage struct {
    ID           uuid.UUID `json:"id"`
    SubmissionID uuid.UUID `json:"submission_id"`
    Data         []byte    `json:"data"`
    CreatedAt    time.Time `json:"created_at"`
}

type Submission struct {
    ID          uuid.UUID       `json:"id"`
    SubmittedBy uuid.UUID       `json:"submitted_by"`
    Category    string          `json:"category"`
    Status      string          `json:"status"`
    Payload     json.RawMessage `json:"payload"`
    PayloadHash string          `json:"payload_hash"`
    ReviewedAt  *time.Time      `json:"reviewed_at"`
    CreatedAt   time.Time       `json:"created_at"`
    DeletedAt   *time.Time      `json:"deleted_at"`
}

type Location struct {
    ID        string     `json:"id"`
    Street    *string    `json:"street"`
    Area      *string    `json:"area"`
    City      *string    `json:"city"`
    Country   *string    `json:"country"`
    Zip       *string    `json:"zip"`
    Lat       *float64   `json:"lat"`
    Lng       *float64   `json:"lng"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at"`
}

type VenueChain struct {
    ID        string     `json:"id"`
    Name      string     `json:"name"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at"`
}

type Venue struct {
    ID           string     `json:"id"`
    LocationID   string     `json:"location_id"`
    VenueChainID *string    `json:"venue_chain_id"`
    Name         string     `json:"name"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
    DeletedAt    *time.Time `json:"deleted_at"`
}

type AlcoholicBeverage struct {
    ID          string     `json:"id"`
    Name        string     `json:"name"`
    ABV         *float32   `json:"abv"`
    Description *string    `json:"description"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    DeletedAt   *time.Time `json:"deleted_at"`
}

type Unit struct {
    ID         string     `json:"id"`
    BeverageID string     `json:"beverage_id"`
    Name       string     `json:"name"`
    VolumeML   *int       `json:"volume_ml"`
    Size        *string      `json:"size"`
    UnitType   *string    `json:"unit_type"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
    DeletedAt  *time.Time `json:"deleted_at"`
}

type VenueUnit struct {
    ID        string     `json:"id"`
    VenueID   string     `json:"venue_id"`
    UnitID    string     `json:"unit_id"`
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at"`
}

type PriceRecord struct {
    ID          string     `json:"id"`
    VenueUnitID string     `json:"venue_unit_id"`
    Currency    string     `json:"currency"`
    Amount      float64    `json:"amount"`
    RecordedAt  time.Time  `json:"recorded_at"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    DeletedAt   *time.Time `json:"deleted_at"`
}

type HappyHour struct {
    ID         string     `json:"id"`
    VenueID    string     `json:"venue_id"`
    DayOfWeek  int16      `json:"day_of_week"`
    StartTime  string     `json:"start_time"` 
    EndTime    string     `json:"end_time"`
    IsActive   bool       `json:"is_active"`
    Name       *string    `json:"name"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
    DeletedAt  *time.Time `json:"deleted_at"`
}

type BusinessHours struct {
    ID         string     `json:"id"`
    VenueID    string     `json:"venue_id"`
    DayOfWeek  int16      `json:"day_of_week"`
    OpenTime   *string    `json:"open_time"` 
    CloseTime  *string    `json:"close_time"`
    IsClosed   bool       `json:"is_closed"`
    CreatedAt  time.Time  `json:"created_at"`
    UpdatedAt  time.Time  `json:"updated_at"`
    DeletedAt  *time.Time `json:"deleted_at"`
}

type User struct {
    ID          string     `json:"id"`
    AuthID      *string    `json:"auth_id"`
    Email       *string    `json:"email"`
    Username    *string    `json:"username"`
    DisplayName *string    `json:"display_name"`
    AvatarURL   *string    `json:"avatar_url"`
    Bio         *string    `json:"bio"`
    DateOfBirth *time.Time `json:"date_of_birth"`
    Phone       *string    `json:"phone"`
    IsVerified  bool       `json:"is_verified"`
    LastLoginAt *time.Time `json:"last_login_at"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
    DeletedAt   *time.Time `json:"deleted_at"`
}