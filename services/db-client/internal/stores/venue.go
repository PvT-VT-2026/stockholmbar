package stores

import (
	"context"
	"database/sql"
	"db-client/internal/db"
	"db-client/internal/models"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type VenueStore struct {
	db *db.DBClient
}

func NewVenueStore(db *db.DBClient) *VenueStore {
	return &VenueStore{db:db}
}

func (s *VenueStore) Create(ctx context.Context, input *models.CreateVenuePayload) (error) {

	// This method inserts into multiple tables (venue & location). Instead of doing two separate insertions, we wrap them in a transaction
	// so that if one fails, the other one rolls back as well. 
    tx, err := s.db.DB().BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("Create: begin tx: %w", err)
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
        return fmt.Errorf("Create: insert location: %w", err)
    }

	// Insert the venue
    _, err = tx.ExecContext(ctx, `
        INSERT INTO venue (location_id, venue_chain_id, name)
        VALUES ($1, $2, $3)
    `,
        locationID,
        input.VenueChainID,
        input.Name,
    )
    if err != nil {
        return fmt.Errorf("Create: insert venue: %w", err)
    }

	// Commit both insertions
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("Create: commit: %w", err)
    }

	return nil
}

// Returns a GetVenueByIDResponse, which contains all the data from the venue table as well as the location data
func (s *VenueStore) GetByID (ctx context.Context, id uuid.UUID) (*models.GetVenueByIDResponse, error) {
    var venue models.GetVenueByIDResponse
    var location models.Location

    err := s.db.DB().QueryRowContext(ctx, `
        SELECT 
            v.id, 
            v.name, 
            v.created_at, 
            v.updated_at, 
		    l.id, 
            l.street, 
            l.area, 
            l.city, 
            l.country, 
            l.zip, 
            l.lat, 
            l.lng, 
            l.created_at, 
            l.updated_at
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
    if errors.Is(err, sql.ErrNoRows) {
        return nil, nil
    }
    if err != nil {
        return nil, fmt.Errorf("GetByID: %w", err)
    }

    venue.Location = location
    return &venue, nil
}

type VenueListFilter struct {
    Category *string
    BeverageNames *[]string
    MaxPrice *int
    Time time.Time
    OnlyHappyHour bool
}


func (s *VenueStore) List(ctx context.Context, filter VenueListFilter) (*models.FilterVenuesResponse, error) {

    // Golang developers chose to represent mon-sat as 1-6, and sunday as 0 instead of 7.
    // We can choose to go with this convention, or set weekday to 7 if it is 0. 
    // However that would require us to do that every time we deal with days, so it is simpler to allow sunday to be represented as 0.
    weekdayInt := int(time.Now().Weekday())
    timeOfDay := filter.Time.Format("15:04") 

    // Business hour filter
    // Only display bars that are open at the given time
    businessHourFilter := fmt.Sprintf(`
        JOIN business_hours bh
        ON bh.venue_id = v.id
        AND bh.day_of_week = %d
        AND (
            (bh.open_time <= '%s'::time AND bh.close_time >= '%s'::time)
        OR
            (bh.open_time > bh.close_time  -- overnight case
            AND ('%s'::time >= bh.open_time OR '%s'::time <= bh.close_time)))`, 
            weekdayInt, timeOfDay, timeOfDay, timeOfDay, timeOfDay,
    )

    // Always LEFT JOIN happy_hours to get the data,
    // but filter on it only if OnlyHappyHour is set
    happyHourJoin := fmt.Sprintf(`
        LEFT JOIN happy_hours hh
        ON hh.venue_id = v.id
        AND hh.day_of_week = %d
        AND (
            (hh.start_time <= '%s'::time AND hh.end_time >= '%s'::time)
        OR
            (hh.start_time > hh.end_time
            AND ('%s'::time >= hh.start_time OR '%s'::time <= hh.end_time)))`,
        weekdayInt, timeOfDay, timeOfDay, timeOfDay, timeOfDay,
    )

    // base query
    query := fmt.Sprintf(`
        SELECT DISTINCT ON (v.id)
            v.id AS venue_id,
            v.name AS venue_name,

            vc.name AS chain_name,
            
            l.street,
            l.area,
            l.city,
            l.country,
            l.zip,
            l.lat,
            l.lng,
            
            u.unit_type,
            u.name AS unit_name,
            u.volume_ml,
            u.size,
            
            pr.amount,
            
            bh.open_time::text,
            bh.close_time::text,
            
            hh.start_time::text AS happy_hour_start,
            hh.end_time::text AS happy_hour_end,
            
            be.name AS beverage_name,
            be.description,
            be.abv
        FROM venue v
        JOIN location l ON l.id = v.location_id
        JOIN venue_unit vu ON vu.venue_id = v.id
        JOIN unit u ON u.id = vu.unit_id
        LEFT JOIN price_record pr ON pr.venue_unit_id = vu.id
        JOIN beverage be ON be.id = u.beverage_id
        LEFT JOIN venue_chain vc ON vc.id = v.venue_chain_id
        %s
        %s
        `, businessHourFilter, happyHourJoin)


    conditions := []string{"v.deleted_at IS NULL"}

    // Category filter
    // Only display bars that have associated units with for example category = "beer"
    if filter.Category != nil {
        conditions = append(conditions, fmt.Sprintf("u.unit_type = '%s'", *filter.Category))
    }

    // Beverage names filter
    if filter.BeverageNames != nil && len(*filter.BeverageNames) > 0 {
        quoted := []string{}
        for _, name := range *filter.BeverageNames {
            quoted = append(quoted,  fmt.Sprintf("'%s'", name))
        }
        fmt.Println("DEBUG: ", quoted)
        conditions = append(conditions, fmt.Sprintf("be.name IN (%s)", strings.Join(quoted, ",")))
    }

    // Max price filter
    // Only display units that have units cheaper than filter.MaxPrice
    if filter.MaxPrice != nil {
        conditions = append(conditions, fmt.Sprintf("pr.amount <= %d", *filter.MaxPrice)) 
    }

    // Happy hour filter
    // Only display bars that have happy hour at the given time
    if filter.OnlyHappyHour {
        conditions = append(conditions, `hh.id IS NOT NULL`)
    }

    query += "\nWHERE " + strings.Join(conditions, "\nAND ")
    query += "\nORDER BY v.id, pr.amount ASC"
    
    fmt.Println(query)

    rows, err := s.db.DB().QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("VenueStore.List: %w", err)
    }
    defer rows.Close()

    return createFilterVenuesResponse(rows)
}



func createFilterVenuesResponse(rows *sql.Rows) (*models.FilterVenuesResponse, error) {
    // Use a map to deduplicate venues by ID since we DISTINCT ON v.id in SQL,
    // but ordering by pr.amount means we still only get one row per venue.
    // The map preserves insertion order isn't guaranteed, so track order separately.
    venues := []*models.FilteredVenue{}
    
    for rows.Next() {
        var (
            venueID     string
            venueName   string
            chainName   *string

            street, area, city, country, zip string
            lat, lng                         float64

            unitType  *string
            unitName  string
            volumeML  *int
            size      *string

            amount float64

            openTime  string
            closeTime string

            happyHourStart string
            happyHourEnd   string

            beverageName        string
            beverageDescription *string
            abv                 *float32
        )

        err := rows.Scan(
            &venueID,
            &venueName,
            &chainName,
            &street,
            &area,
            &city,
            &country,
            &zip,
            &lat,
            &lng,
            &unitType,
            &unitName,
            &volumeML,
            &size,
            &amount,
            &openTime,
            &closeTime,
            &happyHourStart,
            &happyHourEnd,
            &beverageName,
            &beverageDescription,
            &abv,
        )
        if err != nil {
            return nil, fmt.Errorf("CreateFilterVenuesResponse scan: %w", err)
        }

        // Build VenueHours
        hasHappyHour := happyHourStart != "" && happyHourEnd != ""
        hours := models.VenueHours{
            OpenTime:    openTime,
            ClosingTime: closeTime,
            HasHappyHour: hasHappyHour,
        }
        if hasHappyHour {
            hours.HappyHourStart = happyHourStart
            hours.HappyHourEnd = happyHourEnd
        }

        venue := &models.FilteredVenue{
            ID:        venueID,
            VenueName: venueName,
            ChainName: chainName,
            Location: models.FilteredVenueLocation{
                Street:  street,
                Area:    area,
                City:    city,
                Country: country,
                Zip:     zip,
                Lat:     lat,
                Lng:     lng,
            },
            Hours: hours,
            MatchedUnit: models.MatchedUnit{
                UnitName:            unitName,
                BeverageName:        beverageName,
                UnitType:            unitType,
                ABV:                 abv,
                VolumeML:            volumeML,
                Size:                size,
                Price:               amount,
                BeverageDescription: beverageDescription,
            },
        }

        venues = append(venues, venue)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("CreateFilterVenuesResponse rows: %w", err)
    }

    return &models.FilterVenuesResponse{Venues: venues}, nil
}