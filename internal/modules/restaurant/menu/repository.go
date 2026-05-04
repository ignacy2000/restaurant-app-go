package menu

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

func (r *Repository) Create(ctx context.Context, m *Menu) error {
	query := `
		INSERT INTO menus (restaurant_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, m.RestaurantID, m.Name, m.Description).
		Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

func (r *Repository) FindByRestaurantID(ctx context.Context, restaurantID string) ([]Menu, error) {
	query := `SELECT id, restaurant_id, name, description, created_at, updated_at FROM menus WHERE restaurant_id = $1 ORDER BY created_at ASC`
	rows, err := r.db.QueryContext(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("find by restaurant id: %w", err)
	}
	defer rows.Close()

	var results []Menu
	for rows.Next() {
		var m Menu
		if err := rows.Scan(&m.ID, &m.RestaurantID, &m.Name, &m.Description, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan menu: %w", err)
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (r *Repository) CreateItem(ctx context.Context, item *MenuItem) error {
	query := `
		INSERT INTO menu_items (menu_id, name, description, price, position)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, item.MenuID, item.Name, item.Description, item.Price, item.Position).
		Scan(&item.ID, &item.CreatedAt)
}

func (r *Repository) FindItemsByRestaurantID(ctx context.Context, restaurantID string) ([]MenuItem, error) {
	query := `
		SELECT mi.id, mi.menu_id, mi.name, mi.description, mi.price, mi.position, mi.created_at
		FROM menu_items mi
		JOIN menus m ON m.id = mi.menu_id
		WHERE m.restaurant_id = $1
		ORDER BY mi.position, mi.created_at`
	rows, err := r.db.QueryContext(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("find items by restaurant: %w", err)
	}
	defer rows.Close()
	var results []MenuItem
	for rows.Next() {
		var item MenuItem
		if err := rows.Scan(&item.ID, &item.MenuID, &item.Name, &item.Description, &item.Price, &item.Position, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan menu item: %w", err)
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

func (r *Repository) FindMenuRestaurantID(ctx context.Context, menuID string) (string, error) {
	var restaurantID string
	err := r.db.QueryRowContext(ctx, `SELECT restaurant_id FROM menus WHERE id = $1`, menuID).Scan(&restaurantID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return restaurantID, err
}

func (r *Repository) FindItemRestaurantID(ctx context.Context, itemID string) (string, error) {
	var restaurantID string
	query := `SELECT m.restaurant_id FROM menu_items mi JOIN menus m ON m.id = mi.menu_id WHERE mi.id = $1`
	err := r.db.QueryRowContext(ctx, query, itemID).Scan(&restaurantID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return restaurantID, err
}

func (r *Repository) DeleteItem(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM menu_items WHERE id = $1`, id)
	return err
}
