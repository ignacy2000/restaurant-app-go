package restaurant

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

func (r *Repository) Create(ctx context.Context, rest *Restaurant) error {
	query := `
		INSERT INTO restaurants (user_id, name, description, address)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, rest.UserID, rest.Name, rest.Description, rest.Address).
		Scan(&rest.ID, &rest.CreatedAt, &rest.UpdatedAt)
}

func (r *Repository) FindByID(ctx context.Context, id string) (*Restaurant, error) {
	rest := &Restaurant{}
	query := `SELECT id, user_id, name, description, address, created_at, updated_at FROM restaurants WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&rest.ID, &rest.UserID, &rest.Name, &rest.Description, &rest.Address, &rest.CreatedAt, &rest.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find by id: %w", err)
	}
	return rest, nil
}

// FindName returns the name of the restaurant, or "" if not found.
func (r *Repository) FindName(ctx context.Context, id string) (string, error) {
	var name string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM restaurants WHERE id = $1`, id).Scan(&name)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("find name: %w", err)
	}
	return name, nil
}

// FindOwnerID returns the user_id of the restaurant owner, or "" if not found.
func (r *Repository) FindOwnerID(ctx context.Context, id string) (string, error) {
	var ownerID string
	err := r.db.QueryRowContext(ctx, `SELECT user_id FROM restaurants WHERE id = $1`, id).Scan(&ownerID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("find owner id: %w", err)
	}
	return ownerID, nil
}

func (r *Repository) FindByUserID(ctx context.Context, userID string) ([]Restaurant, error) {
	query := `SELECT id, user_id, name, description, address, created_at, updated_at FROM restaurants WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("find by user id: %w", err)
	}
	defer rows.Close()

	var results []Restaurant
	for rows.Next() {
		var rest Restaurant
		if err := rows.Scan(&rest.ID, &rest.UserID, &rest.Name, &rest.Description, &rest.Address, &rest.CreatedAt, &rest.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan restaurant: %w", err)
		}
		results = append(results, rest)
	}
	return results, rows.Err()
}
