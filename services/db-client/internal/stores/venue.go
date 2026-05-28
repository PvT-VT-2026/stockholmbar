package stores

import (
	"context"
	"database/sql"
	"db-client/internal/models"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")

type VenueStore struct {
	pool *pgxpool.Pool
}

func NewVenueStore(pool *pgxpool.Pool) *VenueStore {
	return &VenueStore{pool:pool}
}

func (s *VenueStore) Create(ctx context.Context, input *models.CreateVenuePayload) (uuid.UUID, error) {

	// This method inserts into multiple tables (venue & location). Instead of doing two separate insertions, we wrap them in a transaction
	// so that if one fails, the other one rolls back as well.
    tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        return uuid.Nil, fmt.Errorf("Create: begin tx: %w", err)
    }
    defer tx.Rollback(ctx)

	// Insert the location and get the auto generated id.
    var locationID uuid.UUID
    err = tx.QueryRow(ctx, `
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
        return uuid.Nil, fmt.Errorf("Create: insert location: %w", err)
    }

	// Insert the venue and return its ID
    var venueID uuid.UUID
    err = tx.QueryRow(ctx, `
        INSERT INTO venue (location_id, venue_chain_id, name)
        VALUES ($1, $2, $3)
        RETURNING id
    `,
        locationID,
        input.VenueChainID,
        input.Name,
    ).Scan(&venueID)
    if err != nil {
        return uuid.Nil, fmt.Errorf("Create: insert venue: %w", err)
    }

	// Commit both insertions
    if err := tx.Commit(ctx); err != nil {
        return uuid.Nil, fmt.Errorf("Create: commit: %w", err)
    }

	return venueID, nil
}

func (s *VenueStore) CreateBusinessHours(ctx context.Context, venueID uuid.UUID, hours []models.BusinessHours) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("CreateBusinessHours: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, h := range hours {
		_, err := tx.Exec(ctx, `
			INSERT INTO business_hours (venue_id, day_of_week, open_time, close_time, is_closed)
			VALUES ($1, $2, $3, $4, $5)
		`, venueID, h.DayOfWeek, h.OpenTime, h.CloseTime, h.IsClosed)
		if err != nil {
			return fmt.Errorf("CreateBusinessHours: insert day %d: %w", h.DayOfWeek, err)
		}
	}

	return tx.Commit(ctx)
}

// Returns a GetVenueByIDResponse, which contains all the data from the venue table as well as the location data
func (s *VenueStore) GetByID (ctx context.Context, id uuid.UUID) (*models.GetVenueByIDResponse, error) {
    var venue models.GetVenueByIDResponse
    var location models.Location

    err := s.pool.QueryRow(ctx, `
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

func (s *VenueStore) GetMenuByVenueID(ctx context.Context, id uuid.UUID) ([]models.VenueMenuItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT vu.id, b.name, b.abv, u.name, u.unit_type, u.volume_ml, u.size,
		       pr.currency, pr.amount, pr.recorded_at
		FROM venue_unit vu
		JOIN unit u ON u.id = vu.unit_id AND u.deleted_at IS NULL
		JOIN beverage b ON b.id = u.beverage_id AND b.deleted_at IS NULL
		LEFT JOIN LATERAL (
		    SELECT currency, amount, recorded_at FROM price_record
		    WHERE venue_unit_id = vu.id AND deleted_at IS NULL
		    ORDER BY recorded_at DESC LIMIT 1
		) pr ON true
		WHERE vu.venue_id = $1 AND vu.deleted_at IS NULL
		ORDER BY b.name, u.volume_ml NULLS LAST
	`, id)
	if err != nil {
		return nil, fmt.Errorf("VenueStore.GetMenuByVenueID: %w", err)
	}
	defer rows.Close()

	items := []models.VenueMenuItem{}
	for rows.Next() {
		var m models.VenueMenuItem
		if err := rows.Scan(
			&m.VenueUnitID, &m.BeverageName, &m.ABV, &m.UnitName,
			&m.UnitType, &m.VolumeMl, &m.Size,
			&m.Currency, &m.Amount, &m.RecordedAt,
		); err != nil {
			return nil, fmt.Errorf("VenueStore.GetMenuByVenueID scan: %w", err)
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

type VenueListFilter struct {
    Category *string
    BeverageNames *[]string
    MaxPrice *int
    Time *time.Time
    OnlyHappyHour bool
}

func (s *VenueStore) List(ctx context.Context, filter VenueListFilter) (*models.FilterVenuesResponse, error) {

    // Golang developers chose to represent mon-sat as 1-6, and sunday as 0 instead of 7.
    // We can choose to go with this convention, or set weekday to 7 if it is 0. 
    // However that would require us to do that every time we deal with days, so it is simpler to allow sunday to be represented as 0.
    weekdayInt := int(time.Now().Weekday())

    var businessHourFilter, happyHourJoin string

    if filter.Time != nil {
        // Filter to venues open at the given time (existing behaviour)
        timeOfDay := filter.Time.Format("15:04")
        businessHourFilter = fmt.Sprintf(`
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
        happyHourJoin = fmt.Sprintf(`
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
    } else {
        // No time provided: return all venues scheduled for today (not marked closed)
        businessHourFilter = fmt.Sprintf(`
        JOIN business_hours bh
        ON bh.venue_id = v.id
        AND bh.day_of_week = %d
        AND bh.is_closed IS NOT TRUE`,
            weekdayInt,
        )
        happyHourJoin = fmt.Sprintf(`
        LEFT JOIN happy_hours hh
        ON hh.venue_id = v.id
        AND hh.day_of_week = %d
        AND hh.is_active IS TRUE`,
            weekdayInt,
        )
    }

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

    rows, err := s.pool.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("VenueStore.List: %w", err)
    }
    defer rows.Close()

    return createFilterVenuesResponse(rows)
}



// SearchAll returns every non-deleted venue with its basic location data.
// An optional name query (case-insensitive) filters the results.
func (s *VenueStore) SearchAll(ctx context.Context, query string) ([]models.VenueSearchItem, error) {
	var (
		rows pgx.Rows
		err  error
	)
	if query == "" {
		rows, err = s.pool.Query(ctx, `
			SELECT v.id, v.name,
			       COALESCE(l.street, ''),
			       COALESCE(l.area,   ''),
			       COALESCE(l.city,   '')
			FROM venue v
			JOIN location l ON l.id = v.location_id
			WHERE v.deleted_at IS NULL
			ORDER BY v.name
		`)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT v.id, v.name,
			       COALESCE(l.street, ''),
			       COALESCE(l.area,   ''),
			       COALESCE(l.city,   '')
			FROM venue v
			JOIN location l ON l.id = v.location_id
			WHERE v.deleted_at IS NULL
			  AND v.name ILIKE '%' || $1 || '%'
			ORDER BY v.name
		`, query)
	}
	if err != nil {
		return nil, fmt.Errorf("VenueStore.SearchAll: %w", err)
	}
	defer rows.Close()

	items := []models.VenueSearchItem{}
	for rows.Next() {
		var item models.VenueSearchItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Street, &item.Area, &item.City); err != nil {
			return nil, fmt.Errorf("VenueStore.SearchAll scan: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Update applies partial updates to a venue's name and/or location fields.
func (s *VenueStore) Update(ctx context.Context, venueID uuid.UUID, input models.UpdateVenueInput) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("VenueStore.Update: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if input.Name != nil {
		tag, err := tx.Exec(ctx,
			`UPDATE venue SET name = $1 WHERE id = $2 AND deleted_at IS NULL`,
			*input.Name, venueID,
		)
		if err != nil {
			return fmt.Errorf("VenueStore.Update: update name: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
	}

	locFields := []string{}
	locArgs := []any{}
	idx := 1
	for _, f := range []struct {
		val *string
		col string
	}{
		{input.Street, "street"},
		{input.Area, "area"},
		{input.City, "city"},
		{input.Zip, "zip"},
		{input.Country, "country"},
	} {
		if f.val != nil {
			locFields = append(locFields, fmt.Sprintf("%s = $%d", f.col, idx))
			locArgs = append(locArgs, *f.val)
			idx++
		}
	}
	if len(locFields) > 0 {
		locArgs = append(locArgs, venueID)
		q := fmt.Sprintf(
			`UPDATE location SET %s FROM venue WHERE location.id = venue.location_id AND venue.id = $%d AND venue.deleted_at IS NULL`,
			strings.Join(locFields, ", "), idx,
		)
		if _, err := tx.Exec(ctx, q, locArgs...); err != nil {
			return fmt.Errorf("VenueStore.Update: update location: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// Delete soft-deletes a venue by setting deleted_at.
func (s *VenueStore) Delete(ctx context.Context, venueID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE venue SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`,
		venueID,
	)
	if err != nil {
		return fmt.Errorf("VenueStore.Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateMenuItem applies partial updates to a venue_unit's unit, beverage, and price data.
func (s *VenueStore) UpdateMenuItem(ctx context.Context, venueID, venueUnitID uuid.UUID, input models.UpdateMenuItemInput) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("VenueStore.UpdateMenuItem: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Resolve unit_id and beverage_id, validating ownership.
	var unitID, beverageID uuid.UUID
	err = tx.QueryRow(ctx, `
		SELECT u.id, u.beverage_id
		FROM venue_unit vu
		JOIN unit u ON u.id = vu.unit_id AND u.deleted_at IS NULL
		WHERE vu.id = $1 AND vu.venue_id = $2 AND vu.deleted_at IS NULL
	`, venueUnitID, venueID).Scan(&unitID, &beverageID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("VenueStore.UpdateMenuItem: resolve ids: %w", err)
	}

	// Update unit fields.
	unitFields := []string{}
	unitArgs := []any{}
	idx := 1
	for _, f := range []struct {
		val any
		col string
	}{
		{input.UnitName, "name"},
		{input.UnitType, "unit_type"},
		{input.VolumeMl, "volume_ml"},
	} {
		if f.val != nil {
			unitFields = append(unitFields, fmt.Sprintf("%s = $%d", f.col, idx))
			unitArgs = append(unitArgs, f.val)
			idx++
		}
	}
	if len(unitFields) > 0 {
		unitArgs = append(unitArgs, unitID)
		q := fmt.Sprintf(
			`UPDATE unit SET %s WHERE id = $%d AND deleted_at IS NULL`,
			strings.Join(unitFields, ", "), idx,
		)
		if _, err := tx.Exec(ctx, q, unitArgs...); err != nil {
			return fmt.Errorf("VenueStore.UpdateMenuItem: update unit: %w", err)
		}
	}

	// Update beverage fields.
	bevFields := []string{}
	bevArgs := []any{}
	idx = 1
	for _, f := range []struct {
		val any
		col string
	}{
		{input.BeverageName, "name"},
		{input.ABV, "abv"},
	} {
		if f.val != nil {
			bevFields = append(bevFields, fmt.Sprintf("%s = $%d", f.col, idx))
			bevArgs = append(bevArgs, f.val)
			idx++
		}
	}
	if len(bevFields) > 0 {
		bevArgs = append(bevArgs, beverageID)
		q := fmt.Sprintf(
			`UPDATE beverage SET %s WHERE id = $%d AND deleted_at IS NULL`,
			strings.Join(bevFields, ", "), idx,
		)
		if _, err := tx.Exec(ctx, q, bevArgs...); err != nil {
			return fmt.Errorf("VenueStore.UpdateMenuItem: update beverage: %w", err)
		}
	}

	// Insert new price record when amount is provided (time-series — never mutate existing rows).
	if input.Amount != nil {
		currency := "kr"
		if input.Currency != nil {
			currency = *input.Currency
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO price_record (venue_unit_id, currency, amount) VALUES ($1, $2, $3)`,
			venueUnitID, currency, *input.Amount,
		); err != nil {
			return fmt.Errorf("VenueStore.UpdateMenuItem: insert price_record: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// DeleteMenuItem soft-deletes a single venue_unit row.
func (s *VenueStore) DeleteMenuItem(ctx context.Context, venueID, venueUnitID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE venue_unit SET deleted_at = NOW() WHERE id = $1 AND venue_id = $2 AND deleted_at IS NULL`,
		venueUnitID, venueID,
	)
	if err != nil {
		return fmt.Errorf("VenueStore.DeleteMenuItem: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func createFilterVenuesResponse(rows pgx.Rows) (*models.FilterVenuesResponse, error) {
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

            happyHourStart *sql.NullString
            happyHourEnd   *sql.NullString

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
        hasHappyHour := happyHourStart != nil && happyHourEnd != nil
        hours := models.VenueHours{
            OpenTime:    openTime,
            ClosingTime: closeTime,
            HasHappyHour: hasHappyHour,
        }
        if hasHappyHour {
            hours.HappyHourStart = &happyHourStart.String
            hours.HappyHourEnd = &happyHourEnd.String
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
