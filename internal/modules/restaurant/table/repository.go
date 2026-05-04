package table

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, t *Table) error {
	query := `
		INSERT INTO tables (restaurant_id, number, capacity)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, t.RestaurantID, t.Number, t.Capacity).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *Repository) FindByID(ctx context.Context, id string) (*Table, error) {
	t := &Table{}
	query := `SELECT id, restaurant_id, number, capacity, created_at, updated_at FROM tables WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&t.ID, &t.RestaurantID, &t.Number, &t.Capacity, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find table by id: %w", err)
	}
	return t, nil
}

// FindRestaurantID returns the restaurant_id of the table, or "" if not found.
func (r *Repository) FindRestaurantID(ctx context.Context, id string) (string, error) {
	var restaurantID string
	err := r.db.QueryRowContext(ctx, `SELECT restaurant_id FROM tables WHERE id = $1`, id).Scan(&restaurantID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("find restaurant id: %w", err)
	}
	return restaurantID, nil
}

func (r *Repository) FindByRestaurantID(ctx context.Context, restaurantID string) ([]Table, error) {
	query := `SELECT id, restaurant_id, number, capacity, created_at, updated_at FROM tables WHERE restaurant_id = $1 ORDER BY number ASC`
	rows, err := r.db.QueryContext(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("find tables: %w", err)
	}
	defer rows.Close()

	var result []Table
	for rows.Next() {
		var t Table
		if err := rows.Scan(&t.ID, &t.RestaurantID, &t.Number, &t.Capacity, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

func (r *Repository) UpdateCapacity(ctx context.Context, id string, capacity int) error {
	res, err := r.db.ExecContext(ctx, `UPDATE tables SET capacity = $1 WHERE id = $2`, capacity, id)
	if err != nil {
		return fmt.Errorf("update table: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tables WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete table: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
