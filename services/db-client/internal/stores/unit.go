package stores

import (
	"context"
	"database/sql"
	
	"db-client/internal/models"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitStore struct {
	pool *pgxpool.Pool
}

func NewUnitStore(pool *pgxpool.Pool) *UnitStore {
	return &UnitStore{pool: pool}
}


// This method inserts units, creates venue_unit entries, adds price records and
// creates beverages if a unit does not have a corresponding beverage already.
func (s *UnitStore) Create(ctx context.Context, input *models.CreateUnitsPayload) error {
	// Wrap inserts into a transaction so that if one fails, they will all roll back.
    tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        return fmt.Errorf("CreateUnits: begin tx: %w", err)
    }
    defer tx.Rollback(ctx)

	// Iterate over every unit in the request. 
	// For each unit:
	//	- Check if it has an associated beverage, otherwise create one
	// 	- Insert the unit itself
	//  - Insert a row into venue_unit, which relates a unit (menu item) to a specific venue
	//  - Insert a price record
	for _, unit := range input.Units {
		var beverageID *uuid.UUID
		err := tx.QueryRow(ctx, `SELECT id FROM beverage WHERE name = $1 AND abv = $2`, unit.Name, unit.ABV).Scan(&beverageID)
		if errors.Is(err, sql.ErrNoRows) {
			beverageID, err = s.createBeverage(ctx, tx, unit)
			if err != nil {
				return err
			}
		} else if err != nil {
			return fmt.Errorf("CreateUnits: failed to query beverage: %w", err)
		}

		unitID, err := s.createUnit(ctx, tx, unit, beverageID)
		if err != nil {
			return err
		}
		
		venueUnitID, err := s.createVenueUnit(ctx, tx, input.VenueID, unitID)
		if err != nil {
			return err
		}

		err = s.createPriceRecord(ctx, tx, unit, venueUnitID)
		if err != nil {
			return err
		}
	}

	// Commit the whole transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("CreateUnits: commit failed: %w", err)
	}

	return nil
}	

func (s *UnitStore) createBeverage(ctx context.Context, tx pgx.Tx, unit *models.UnitInput) (*uuid.UUID, error) {
	var beverageID *uuid.UUID
	err := tx.QueryRow(ctx, `
        INSERT INTO beverage (name, abv)
        VALUES ($1, $2)
        ON CONFLICT ON CONSTRAINT beverage_name_abv_unique DO NOTHING
        RETURNING id
    `, unit.Name, unit.ABV).Scan(&beverageID)

	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRow(ctx, `
            SELECT id FROM beverage WHERE name = $1 AND abv = $2
        `, unit.Name, unit.ABV).Scan(&beverageID)
	}
	if err != nil {
		return nil, fmt.Errorf("CreateBeverage: %w", err)
	}
	return beverageID, nil
}

func (s *UnitStore) createUnit(ctx context.Context, tx pgx.Tx, unit *models.UnitInput, beverageID *uuid.UUID) (*uuid.UUID, error) {
	var unitID *uuid.UUID
	err := tx.QueryRow(ctx, `
        INSERT INTO unit (beverage_id, name, volume_ml, size, unit_type)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT ON CONSTRAINT unit_beverage_volume_type_unique DO NOTHING
        RETURNING id
    `, beverageID, unit.Name, unit.VolumeML, unit.Size, unit.UnitType).Scan(&unitID)

	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRow(ctx, `
            SELECT id FROM unit WHERE beverage_id = $1 AND volume_ml = $2 AND unit_type = $3
        `, beverageID, unit.VolumeML, unit.UnitType).Scan(&unitID)
	}
	if err != nil {
		return nil, fmt.Errorf("CreateUnit: %w", err)
	}
	return unitID, nil
}

func (s *UnitStore) createVenueUnit(ctx context.Context, tx pgx.Tx, venueID, unitID *uuid.UUID) (*uuid.UUID, error) {
	var venueUnitID *uuid.UUID
	err := tx.QueryRow(ctx, `
        SELECT id FROM venue_unit WHERE venue_id = $1 AND unit_id = $2
    `, venueID, unitID).Scan(&venueUnitID)

	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRow(ctx, `
            INSERT INTO venue_unit (venue_id, unit_id)
            VALUES ($1, $2)
            RETURNING id
        `, venueID, unitID).Scan(&venueUnitID)
	}
	if err != nil {
		return nil, fmt.Errorf("CreateVenueUnit: %w", err)
	}
	return venueUnitID, nil
}

func (s *UnitStore) createPriceRecord(ctx context.Context, tx pgx.Tx, unit *models.UnitInput, venueUnitID *uuid.UUID)	error {
	_, err := tx.Exec(ctx, `
        INSERT INTO price_record (venue_unit_id, currency, amount)
        VALUES ($1, $2, $3)
    `,
		venueUnitID, unit.Currency, unit.Price,
	)
	if err != nil {
		return fmt.Errorf("CreatePriceRecord: unable to insert new price record: %w", err)
	}

	return nil
}
