package stores

import (
	"context"
	"db-client/internal/db"
	"db-client/internal/models"
	"fmt"

	"github.com/google/uuid"
)

type VenueStore struct {
	db *db.DBClient
}

func NewVenueStore(db *db.DBClient) *VenueStore {
	return &VenueStore{db:db}
}

func (s *VenueStore) Create(ctx context.Context, input models.CreateVenueInput) (*models.Venue, error) {

	// This method inserts into multiple tables (venue & location). Instead of doing two separate insertions, we wrap them in a transaction
	// so that if one fails, the other one rolls back as well. 
    tx, err := s.db.DB().BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("Create: begin tx: %w", err)
    }
    defer tx.Rollback()

	// Insert the location and get the auto generated id.
    var locationID uuid.UUID
    err = tx.QueryRowContext(ctx, `
        INSERT INTO location (street, area, city, country, zip, lat, lng)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    	RETURNING id
	`,
        input.Street,
        input.Area,
        input.City,
        input.Country,
        input.Zip,
        input.Lat,
        input.Lng,
    ).Scan(&locationID)
    if err != nil {
        return nil, fmt.Errorf("Create: insert location: %w", err)
    }

	// Insert the venue and get the auto generated id. 
	var venueID uuid.UUID
    err = tx.QueryRowContext(ctx, `
        INSERT INTO venue (location_id, venue_chain_id, name)
        VALUES ($1, $2, $3)
        RETURNING id
    `,
        locationID,
        input.VenueChainID,
        input.Name,
    ).Scan(&venueID)
    if err != nil {
        return nil, fmt.Errorf("Create: insert venue: %w", err)
    }

	// Commit both insertions
    if err := tx.Commit(); err != nil {
        return nil, fmt.Errorf("Create: commit: %w", err)
    }

	return nil,nil
}


func (s *VenueStore) GetByID (ctx context.Context, id uuid.UUID) (*models.VenueResponse, error) {
    var venue models.VenueResponse
    var location models.Location

    err := s.db.DB().QueryRowContext(ctx, `
        SELECT v.id, v.name, v.created_at, v.updated_at, 
		l.id, l.street, l.area, l.city, l.country, l.zip, l.lat, l.lng, l.created_at, l.updated_at
		FROM venue v
		JOIN location l ON l.id = v.location_id
		WHERE v.id = $1
		AND v.deleted_at IS NULL
    `, id).Scan(
        &venue.ID, &venue.Name, &venue.CreatedAt, &venue.UpdatedAt,
        &location.ID, &location.Street, &location.Area, &location.City,
        &location.Country, &location.Zip, &location.Lat, &location.Lng,
        &location.CreatedAt, &location.UpdatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("GetByID: %w", err)
    }

    venue.Location = location
    return &venue, nil
}

