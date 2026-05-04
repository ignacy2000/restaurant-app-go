package user_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"table-service.pl/internal/modules/user"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestUserRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	repo := user.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("alice@example.com", "hash").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("uid-1", now, now))

	u := &user.User{Email: "alice@example.com", PasswordHash: "hash"}
	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if u.ID != "uid-1" {
		t.Errorf("ID = %q, want uid-1", u.ID)
	}
	if u.CreatedAt != now {
		t.Errorf("CreatedAt mismatch")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUserRepository_Create_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := user.NewRepository(db)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("alice@example.com", "hash").
		WillReturnError(errors.New("db error"))

	u := &user.User{Email: "alice@example.com", PasswordHash: "hash"}
	if err := repo.Create(context.Background(), u); err == nil {
		t.Error("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUserRepository_FindByEmail_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := user.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM users WHERE email`).
		WithArgs("alice@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow("uid-1", "alice@example.com", "hash", now, now))

	got, err := repo.FindByEmail(context.Background(), "alice@example.com")
	if err != nil {
		t.Fatalf("FindByEmail() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected user, got nil")
	}
	if got.ID != "uid-1" || got.Email != "alice@example.com" {
		t.Errorf("got %+v, want ID=uid-1 Email=alice@example.com", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := user.NewRepository(db)

	mock.ExpectQuery(`FROM users WHERE email`).
		WithArgs("nobody@example.com").
		WillReturnError(sql.ErrNoRows)

	got, err := repo.FindByEmail(context.Background(), "nobody@example.com")
	if err != nil {
		t.Fatalf("FindByEmail() error = %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUserRepository_FindByEmail_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := user.NewRepository(db)

	mock.ExpectQuery(`FROM users WHERE email`).
		WithArgs("alice@example.com").
		WillReturnError(errors.New("connection reset"))

	_, err := repo.FindByEmail(context.Background(), "alice@example.com")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUserRepository_FindByID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := user.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM users WHERE id`).
		WithArgs("uid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow("uid-1", "alice@example.com", "hash", now, now))

	got, err := repo.FindByID(context.Background(), "uid-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected user, got nil")
	}
	if got.ID != "uid-1" {
		t.Errorf("ID = %q, want uid-1", got.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := user.NewRepository(db)

	mock.ExpectQuery(`FROM users WHERE id`).
		WithArgs("uid-x").
		WillReturnError(sql.ErrNoRows)

	got, err := repo.FindByID(context.Background(), "uid-x")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestUserRepository_FindByID_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := user.NewRepository(db)

	mock.ExpectQuery(`FROM users WHERE id`).
		WithArgs("uid-1").
		WillReturnError(errors.New("timeout"))

	_, err := repo.FindByID(context.Background(), "uid-1")
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
