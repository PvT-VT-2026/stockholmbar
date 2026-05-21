package stores

import (
	"context"
	"db-client/internal/models"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BeverageStore struct {
	pool *pgxpool.Pool
}

func NewBeverageStore(pool *pgxpool.Pool) *BeverageStore {
	return &BeverageStore{pool:pool}
}


func (s *BeverageStore) List(ctx context.Context) (*models.BeverageListResponse, error) {
	rows, err := s.pool.Query(ctx, `
        SELECT DISTINCT
        b.name,
        u.unit_type AS category
        FROM public.beverage b
        JOIN public.unit u
        ON u.beverage_id = b.id
        WHERE b.deleted_at IS NULL
        AND u.deleted_at IS NULL
        ORDER BY b.name, u.unit_type;
	`)
	if err != nil {
		return nil, fmt.Errorf("BeverageStore.List: %w", err)
	}
	defer rows.Close()

	items := []models.BeverageWithCategory{}
	for rows.Next() {
		var b models.BeverageWithCategory
		if err := rows.Scan(
			&b.Name, &b.Category,
		); err != nil {
			return nil, fmt.Errorf("BeverageStore.List scan: %w", err)
		}
		items = append(items, b)
	}

	return &models.BeverageListResponse{Beverages: items}, rows.Err()
}