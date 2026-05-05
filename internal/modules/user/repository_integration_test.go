//go:build integration

package user_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"table-service.pl/internal/modules/user"
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

func TestUserRepository_Create_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	repo := user.NewRepository(integrationDB)

	u := &user.User{Email: "alice@example.com", PasswordHash: "$2a$10$testhash"}
	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if u.ID == "" {
		t.Error("ID should be set by RETURNING")
	}
	if u.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set by RETURNING")
	}
}

func TestUserRepository_FindByEmail_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	repo := user.NewRepository(integrationDB)

	original := &user.User{Email: "bob@example.com", PasswordHash: "hash"}
	if err := repo.Create(context.Background(), original); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.FindByEmail(context.Background(), "bob@example.com")
	if err != nil {
		t.Fatalf("FindByEmail() error = %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.ID != original.ID || found.Email != "bob@example.com" {
		t.Errorf("got %+v", found)
	}

	missing, err := repo.FindByEmail(context.Background(), "nobody@example.com")
	if err != nil {
		t.Fatalf("FindByEmail(missing) error = %v", err)
	}
	if missing != nil {
		t.Errorf("expected nil for unknown email, got %+v", missing)
	}
}

func TestUserRepository_FindByID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	repo := user.NewRepository(integrationDB)

	u := &user.User{Email: "carol@example.com", PasswordHash: "hash"}
	if err := repo.Create(context.Background(), u); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.FindByID(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found == nil || found.Email != "carol@example.com" {
		t.Errorf("got %+v", found)
	}

	missing, err := repo.FindByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("FindByID(missing) error = %v", err)
	}
	if missing != nil {
		t.Errorf("expected nil for unknown ID, got %+v", missing)
	}
}

func TestUserRepository_Create_DuplicateEmail_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	repo := user.NewRepository(integrationDB)

	u1 := &user.User{Email: "dup@example.com", PasswordHash: "h1"}
	if err := repo.Create(context.Background(), u1); err != nil {
		t.Fatalf("Create first: %v", err)
	}

	u2 := &user.User{Email: "dup@example.com", PasswordHash: "h2"}
	if err := repo.Create(context.Background(), u2); err == nil {
		t.Error("expected unique constraint violation, got nil")
	}
}
