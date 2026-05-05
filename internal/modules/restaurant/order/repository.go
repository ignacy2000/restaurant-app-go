package order

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

func (r *Repository) Create(ctx context.Context, o *Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	query := `
		INSERT INTO orders (restaurant_id, table_id, status, notes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	err = tx.QueryRowContext(ctx, query, o.RestaurantID, o.TableID, o.Status, o.Notes).
		Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("insert order: %w", err)
	}

	for i := range o.Items {
		item := &o.Items[i]
		itemQ := `INSERT INTO order_items (order_id, name, quantity, notes) VALUES ($1, $2, $3, $4) RETURNING id`
		if err := tx.QueryRowContext(ctx, itemQ, o.ID, item.Name, item.Quantity, item.Notes).Scan(&item.ID); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert order item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *Repository) FindByRestaurantID(ctx context.Context, restaurantID string) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, restaurant_id, table_id, status, notes, created_at, updated_at FROM orders WHERE restaurant_id = $1 ORDER BY created_at DESC`,
		restaurantID)
	if err != nil {
		return nil, fmt.Errorf("find orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.RestaurantID, &o.TableID, &o.Status, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range orders {
		if err := r.loadItems(ctx, &orders[i]); err != nil {
			return nil, err
		}
	}
	return orders, nil
}

func (r *Repository) FindByTableID(ctx context.Context, restaurantID, tableID string) ([]Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, restaurant_id, table_id, status, notes, created_at, updated_at FROM orders WHERE restaurant_id = $1 AND table_id = $2 ORDER BY created_at DESC`,
		restaurantID, tableID)
	if err != nil {
		return nil, fmt.Errorf("find orders by table: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.RestaurantID, &o.TableID, &o.Status, &o.Notes, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range orders {
		if err := r.loadItems(ctx, &orders[i]); err != nil {
			return nil, err
		}
	}
	return orders, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*Order, error) {
	o := &Order{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, restaurant_id, table_id, status, notes, created_at, updated_at FROM orders WHERE id = $1`, id).
		Scan(&o.ID, &o.RestaurantID, &o.TableID, &o.Status, &o.Notes, &o.CreatedAt, &o.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find order by id: %w", err)
	}
	if err := r.loadItems(ctx, o); err != nil {
		return nil, err
	}
	return o, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id string, status OrderStatus) error {
	res, err := r.db.ExecContext(ctx, `UPDATE orders SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) loadItems(ctx context.Context, o *Order) error {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, order_id, name, quantity, notes FROM order_items WHERE order_id = $1`, o.ID)
	if err != nil {
		return fmt.Errorf("load items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var it OrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.Name, &it.Quantity, &it.Notes); err != nil {
			return fmt.Errorf("scan item: %w", err)
		}
		o.Items = append(o.Items, it)
	}
	return rows.Err()
}
