package call

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

func (r *Repository) Create(ctx context.Context, c *WaiterCall) error {
	query := `
		INSERT INTO waiter_calls (restaurant_id, table_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, c.RestaurantID, c.TableID, c.Status).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *Repository) FindByID(ctx context.Context, id string) (*WaiterCall, error) {
	c := &WaiterCall{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, restaurant_id, table_id, status, created_at, updated_at FROM waiter_calls WHERE id = $1`, id).
		Scan(&c.ID, &c.RestaurantID, &c.TableID, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find call by id: %w", err)
	}
	return c, nil
}

func (r *Repository) FindByRestaurantID(ctx context.Context, restaurantID string) ([]WaiterCall, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, restaurant_id, table_id, status, created_at, updated_at FROM waiter_calls WHERE restaurant_id = $1 ORDER BY created_at DESC`,
		restaurantID)
	if err != nil {
		return nil, fmt.Errorf("find calls: %w", err)
	}
	defer rows.Close()

	var result []WaiterCall
	for rows.Next() {
		var c WaiterCall
		if err := rows.Scan(&c.ID, &c.RestaurantID, &c.TableID, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan call: %w", err)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *Repository) FindActiveByRestaurantID(ctx context.Context, restaurantID string) ([]WaiterCall, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, restaurant_id, table_id, status, created_at, updated_at FROM waiter_calls WHERE restaurant_id = $1 AND status != 'done' ORDER BY created_at ASC`,
		restaurantID)
	if err != nil {
		return nil, fmt.Errorf("find active calls: %w", err)
	}
	defer rows.Close()

	var result []WaiterCall
	for rows.Next() {
		var c WaiterCall
		if err := rows.Scan(&c.ID, &c.RestaurantID, &c.TableID, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan call: %w", err)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status CallStatus) error {
	res, err := r.db.ExecContext(ctx, `UPDATE waiter_calls SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
