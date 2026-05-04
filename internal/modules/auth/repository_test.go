package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"table-service.pl/internal/modules/auth"
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

// --- sessions ---

func TestAuthRepository_CreateSession(t *testing.T) {
	db, mock := newMockDB(t)
	repo := auth.NewRepository(db)
	expires := time.Now().Add(24 * time.Hour)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO sessions`).
		WithArgs("uid-1", "tok-abc", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("sess-1", now))

	s := &auth.Session{UserID: "uid-1", RefreshToken: "tok-abc", ExpiresAt: expires}
	if err := repo.CreateSession(context.Background(), s); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if s.ID != "sess-1" {
		t.Errorf("ID = %q, want sess-1", s.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAuthRepository_CreateSession_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := auth.NewRepository(db)

	mock.ExpectQuery(`INSERT INTO sessions`).
		WillReturnError(errors.New("db error"))

	s := &auth.Session{UserID: "uid-1", RefreshToken: "tok", ExpiresAt: time.Now()}
	if err := repo.CreateSession(context.Background(), s); err == nil {
		t.Error("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAuthRepository_FindByRefreshToken_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := auth.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM sessions WHERE refresh_token`).
		WithArgs("tok-abc").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "refresh_token", "expires_at", "created_at"}).
			AddRow("sess-1", "uid-1", "tok-abc", now.Add(24*time.Hour), now))

	got, err := repo.FindByRefreshToken(context.Background(), "tok-abc")
	if err != nil {
		t.Fatalf("FindByRefreshToken() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected session, got nil")
	}
	if got.ID != "sess-1" || got.UserID != "uid-1" {
		t.Errorf("got %+v, want ID=sess-1 UserID=uid-1", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAuthRepository_FindByRefreshToken_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := auth.NewRepository(db)

	mock.ExpectQuery(`FROM sessions WHERE refresh_token`).
		WithArgs("unknown").
		WillReturnError(sql.ErrNoRows)

	got, err := repo.FindByRefreshToken(context.Background(), "unknown")
	if err != nil {
		t.Fatalf("FindByRefreshToken() error = %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAuthRepository_DeleteByRefreshToken(t *testing.T) {
	db, mock := newMockDB(t)
	repo := auth.NewRepository(db)

	mock.ExpectExec(`DELETE FROM sessions WHERE refresh_token`).
		WithArgs("tok-abc").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.DeleteByRefreshToken(context.Background(), "tok-abc"); err != nil {
		t.Fatalf("DeleteByRefreshToken() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAuthRepository_DeleteByRefreshToken_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := auth.NewRepository(db)

	mock.ExpectExec(`DELETE FROM sessions WHERE refresh_token`).
		WillReturnError(errors.New("db error"))

	if err := repo.DeleteByRefreshToken(context.Background(), "tok"); err == nil {
		t.Error("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// --- allowed origins ---

func TestAuthRepository_AllOriginValues_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	repo := auth.NewRepository(db)

	mock.ExpectQuery(`SELECT origin FROM allowed_origins`).
		WillReturnRows(sqlmock.NewRows([]string{"origin"}))

	origins, err := repo.AllOriginValues(context.Background())
	if err != nil {
		t.Fatalf("AllOriginValues() error = %v", err)
	}
	if len(origins) != 0 {
		t.Errorf("expected empty slice, got %v", origins)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAuthRepository_AllOriginValues_Multiple(t *testing.T) {
	db, mock := newMockDB(t)
	repo := auth.NewRepository(db)

	mock.ExpectQuery(`SELECT origin FROM allowed_origins`).
		WillReturnRows(sqlmock.NewRows([]string{"origin"}).
			AddRow("https://app.example.com").
			AddRow("https://admin.example.com"))

	origins, err := repo.AllOriginValues(context.Background())
	if err != nil {
		t.Fatalf("AllOriginValues() error = %v", err)
	}
	if len(origins) != 2 {
		t.Fatalf("len = %d, want 2", len(origins))
	}
	if origins[0] != "https://app.example.com" {
		t.Errorf("origins[0] = %q", origins[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAuthRepository_ListOrigins(t *testing.T) {
	db, mock := newMockDB(t)
	repo := auth.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`SELECT id, origin, created_at FROM allowed_origins`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "origin", "created_at"}).
			AddRow("oid-1", "https://app.example.com", now))

	list, err := repo.ListOrigins(context.Background())
	if err != nil {
		t.Fatalf("ListOrigins() error = %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len = %d, want 1", len(list))
	}
	if list[0].ID != "oid-1" {
		t.Errorf("ID = %q, want oid-1", list[0].ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAuthRepository_AddOrigin(t *testing.T) {
	db, mock := newMockDB(t)
	repo := auth.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO allowed_origins`).
		WithArgs("https://new.example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "origin", "created_at"}).
			AddRow("oid-2", "https://new.example.com", now))

	got, err := repo.AddOrigin(context.Background(), "https://new.example.com")
	if err != nil {
		t.Fatalf("AddOrigin() error = %v", err)
	}
	if got.ID != "oid-2" {
		t.Errorf("ID = %q, want oid-2", got.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAuthRepository_RemoveOrigin(t *testing.T) {
	db, mock := newMockDB(t)
	repo := auth.NewRepository(db)

	mock.ExpectExec(`DELETE FROM allowed_origins`).
		WithArgs("oid-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.RemoveOrigin(context.Background(), "oid-1"); err != nil {
		t.Fatalf("RemoveOrigin() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestAuthRepository_RemoveOrigin_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := auth.NewRepository(db)

	mock.ExpectExec(`DELETE FROM allowed_origins`).
		WithArgs("oid-x").
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	if err := repo.RemoveOrigin(context.Background(), "oid-x"); err != sql.ErrNoRows {
		t.Errorf("RemoveOrigin() error = %v, want sql.ErrNoRows", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
