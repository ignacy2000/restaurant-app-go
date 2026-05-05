package auth

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

// --- sessions ---

func (r *Repository) CreateSession(ctx context.Context, s *Session) error {
	query := `
		INSERT INTO sessions (user_id, refresh_token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, s.UserID, s.RefreshToken, s.ExpiresAt).
		Scan(&s.ID, &s.CreatedAt)
}

func (r *Repository) FindByRefreshToken(ctx context.Context, token string) (*Session, error) {
	s := &Session{}
	query := `SELECT id, user_id, refresh_token, expires_at, created_at FROM sessions WHERE refresh_token = $1`
	err := r.db.QueryRowContext(ctx, query, token).
		Scan(&s.ID, &s.UserID, &s.RefreshToken, &s.ExpiresAt, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	return s, nil
}

func (r *Repository) DeleteByRefreshToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE refresh_token = $1`, token)
	return err
}

// --- allowed origins ---

func (r *Repository) AllOriginValues(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT origin FROM allowed_origins`)
	if err != nil {
		return nil, fmt.Errorf("query origins: %w", err)
	}
	defer rows.Close()

	var origins []string
	for rows.Next() {
		var o string
		if err := rows.Scan(&o); err != nil {
			return nil, fmt.Errorf("scan origin: %w", err)
		}
		origins = append(origins, o)
	}
	return origins, rows.Err()
}

func (r *Repository) ListOrigins(ctx context.Context) ([]AllowedOrigin, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, origin, created_at FROM allowed_origins ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("list origins: %w", err)
	}
	defer rows.Close()

	var result []AllowedOrigin
	for rows.Next() {
		var o AllowedOrigin
		if err := rows.Scan(&o.ID, &o.Origin, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan origin: %w", err)
		}
		result = append(result, o)
	}
	return result, rows.Err()
}

func (r *Repository) AddOrigin(ctx context.Context, origin string) (*AllowedOrigin, error) {
	o := &AllowedOrigin{}
	query := `INSERT INTO allowed_origins (origin) VALUES ($1) RETURNING id, origin, created_at`
	err := r.db.QueryRowContext(ctx, query, origin).Scan(&o.ID, &o.Origin, &o.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("add origin: %w", err)
	}
	return o, nil
}

func (r *Repository) RemoveOrigin(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM allowed_origins WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("remove origin: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// --- password reset tokens ---

func (r *Repository) CreateResetToken(ctx context.Context, t *PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, t.UserID, t.Token, t.ExpiresAt).
		Scan(&t.ID, &t.CreatedAt)
}

func (r *Repository) FindResetToken(ctx context.Context, token string) (*PasswordResetToken, error) {
	t := &PasswordResetToken{}
	query := `SELECT id, user_id, token, expires_at, created_at FROM password_reset_tokens WHERE token = $1`
	err := r.db.QueryRowContext(ctx, query, token).
		Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find reset token: %w", err)
	}
	return t, nil
}

func (r *Repository) DeleteResetToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE token = $1`, token)
	return err
}
