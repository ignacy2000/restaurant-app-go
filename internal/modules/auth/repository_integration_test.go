//go:build integration

package auth_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"table-service.pl/internal/modules/auth"
	"table-service.pl/internal/testhelper"
)

var integrationDB *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()
	db, cleanup := testhelper.NewPostgresDB(ctx)
	integrationDB = db
	code := m.Run()
	cleanup()
	os.Exit(code)
}

func TestAuthRepository_CreateSession_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "alice@example.com")
	repo := auth.NewRepository(integrationDB)

	s := &auth.Session{
		UserID:       userID,
		RefreshToken: "tok-abc",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	if err := repo.CreateSession(context.Background(), s); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if s.ID == "" {
		t.Error("ID should be populated by RETURNING")
	}
	if s.CreatedAt.IsZero() {
		t.Error("CreatedAt should be populated by RETURNING")
	}
}

func TestAuthRepository_FindByRefreshToken_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "bob@example.com")
	repo := auth.NewRepository(integrationDB)

	s := &auth.Session{UserID: userID, RefreshToken: "tok-xyz", ExpiresAt: time.Now().Add(time.Hour)}
	if err := repo.CreateSession(context.Background(), s); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	found, err := repo.FindByRefreshToken(context.Background(), "tok-xyz")
	if err != nil {
		t.Fatalf("FindByRefreshToken() error = %v", err)
	}
	if found == nil || found.UserID != userID {
		t.Errorf("got %+v, want UserID=%s", found, userID)
	}

	missing, err := repo.FindByRefreshToken(context.Background(), "unknown-tok")
	if err != nil {
		t.Fatalf("FindByRefreshToken(missing) error = %v", err)
	}
	if missing != nil {
		t.Errorf("expected nil for unknown token, got %+v", missing)
	}
}

func TestAuthRepository_DeleteByRefreshToken_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "carol@example.com")
	repo := auth.NewRepository(integrationDB)

	s := &auth.Session{UserID: userID, RefreshToken: "del-tok", ExpiresAt: time.Now().Add(time.Hour)}
	if err := repo.CreateSession(context.Background(), s); err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if err := repo.DeleteByRefreshToken(context.Background(), "del-tok"); err != nil {
		t.Fatalf("DeleteByRefreshToken() error = %v", err)
	}

	found, _ := repo.FindByRefreshToken(context.Background(), "del-tok")
	if found != nil {
		t.Error("expected session to be deleted, still found")
	}
}

func TestAuthRepository_Origins_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	repo := auth.NewRepository(integrationDB)

	// empty state
	vals, err := repo.AllOriginValues(context.Background())
	if err != nil {
		t.Fatalf("AllOriginValues() error = %v", err)
	}
	if len(vals) != 0 {
		t.Errorf("expected empty, got %v", vals)
	}

	// add two origins
	o1, err := repo.AddOrigin(context.Background(), "https://app.example.com")
	if err != nil {
		t.Fatalf("AddOrigin() error = %v", err)
	}
	if o1.ID == "" || o1.Origin != "https://app.example.com" {
		t.Errorf("unexpected AddOrigin result: %+v", o1)
	}

	_, err = repo.AddOrigin(context.Background(), "https://admin.example.com")
	if err != nil {
		t.Fatalf("AddOrigin second error = %v", err)
	}

	// AllOriginValues should return both
	vals, err = repo.AllOriginValues(context.Background())
	if err != nil {
		t.Fatalf("AllOriginValues() error = %v", err)
	}
	if len(vals) != 2 {
		t.Fatalf("AllOriginValues len = %d, want 2", len(vals))
	}

	// ListOrigins
	list, err := repo.ListOrigins(context.Background())
	if err != nil {
		t.Fatalf("ListOrigins() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("ListOrigins len = %d, want 2", len(list))
	}

	// RemoveOrigin
	if err := repo.RemoveOrigin(context.Background(), o1.ID); err != nil {
		t.Fatalf("RemoveOrigin() error = %v", err)
	}

	vals, _ = repo.AllOriginValues(context.Background())
	if len(vals) != 1 {
		t.Fatalf("after remove: len = %d, want 1", len(vals))
	}

	// RemoveOrigin non-existent
	if err := repo.RemoveOrigin(context.Background(), "00000000-0000-0000-0000-000000000000"); err != sql.ErrNoRows {
		t.Errorf("RemoveOrigin(missing) error = %v, want sql.ErrNoRows", err)
	}
}
