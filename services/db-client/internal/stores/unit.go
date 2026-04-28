package stores

import (
	"context"
	"database/sql"
	"db-client/internal/db"
	"db-client/internal/models"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type UnitStore struct {
	db *db.DBClient
}

func NewUnitStore(db *db.DBClient) *UnitStore {
	return &UnitStore{db: db}
}


// This method inserts units, creates venue_unit entries, adds price records and
// creates beverages if a unit does not have a corresponding beverage already.
func (s *UnitStore) CreateUnits(ctx context.Context, input models.CreateUnitsRequest) error {
	// Wrap inserts into a transaction so that if one fails, they will all roll back.
    tx, err := s.db.DB().BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("CreateUnits: begin tx: %w", err)
    }
    defer tx.Rollback()

	// Iterate over every unit in the request. 
	// For each unit:
	//	- Check if it has an associated beverage, otherwise create one
	// 	- Insert the unit itself
	//  - Insert a row into venue_unit, which relates a unit (menu item) to a specific venue
	//  - Insert a price record
	for _, unit := range input.Units {
		var beverageID *uuid.UUID
		err := tx.QueryRowContext(ctx, `SELECT id FROM beverage WHERE name = $1 AND abv = $2`, unit.Name, unit.ABV).Scan(&beverageID)
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
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("CreateUnits: commit failed: %w", err)
	}

	return nil
}	

func (s *UnitStore) createBeverage(ctx context.Context, tx *sql.Tx, unit *models.UnitInput) (*uuid.UUID, error) {
	var beverageID *uuid.UUID
	err := tx.QueryRowContext(ctx, `
        INSERT INTO beverage (name, abv)
        VALUES ($1, $2)
        RETURNING id
    `,
		unit.Name,
		unit.ABV,
	).Scan(&beverageID)

	if err != nil {
		return nil, fmt.Errorf("CreateBeverage: unable to insert new beverage: %w", err)
	}

	return beverageID, nil
}

func (s *UnitStore) createUnit(ctx context.Context, tx *sql.Tx, unit *models.UnitInput, beverageID *uuid.UUID)	(*uuid.UUID, error) {
	var UnitID *uuid.UUID
	err := tx.QueryRowContext(ctx, `
        INSERT INTO unit (beverage_id, name, volume_ml, size, unit_type)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `,
		beverageID, unit.Name, unit.VolumeML, unit.Size, unit.UnitType,
	).Scan(&UnitID)
	if err != nil {
		return nil, fmt.Errorf("CreateUnit: unable to insert new unit: %w", err)
	}

	return UnitID, nil
}

func (s *UnitStore) createVenueUnit(ctx context.Context, tx *sql.Tx, venueID, unitID *uuid.UUID) (*uuid.UUID, error) {
	var venueUnitID *uuid.UUID
	err := tx.QueryRowContext(ctx, `
        INSERT INTO venue_unit (venue_id, unit_id)
        VALUES ($1, $2)
		RETURNING id
		`,
		venueID, unitID,
	).Scan(&venueUnitID)
	if err != nil {
		return nil, fmt.Errorf("CreateVenueUnit: unable to insert new venue unit: %w", err)
	}

	return venueUnitID, nil
}

func (s *UnitStore) createPriceRecord(ctx context.Context, tx *sql.Tx, unit *models.UnitInput, venueUnitID *uuid.UUID)	error {
	_, err := tx.ExecContext(ctx, `
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
