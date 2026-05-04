//go:build integration

package restaurant_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"table-service.pl/internal/modules/restaurant"
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

func TestRestaurantRepository_Create_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	repo := restaurant.NewRepository(integrationDB)

	r := &restaurant.Restaurant{
		UserID:      userID,
		Name:        "La Trattoria",
		Description: "Italian food",
		Address:     "Via Roma 1",
	}
	if err := repo.Create(context.Background(), r); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if r.ID == "" {
		t.Error("ID should be set by RETURNING")
	}
	if r.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set by RETURNING")
	}
}

func TestRestaurantRepository_FindByID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	repo := restaurant.NewRepository(integrationDB)

	original := &restaurant.Restaurant{UserID: userID, Name: "La Trattoria", Description: "D", Address: "A"}
	if err := repo.Create(context.Background(), original); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.FindByID(context.Background(), original.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found == nil || found.Name != "La Trattoria" || found.UserID != userID {
		t.Errorf("got %+v", found)
	}

	missing, err := repo.FindByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("FindByID(missing) error = %v", err)
	}
	if missing != nil {
		t.Errorf("expected nil, got %+v", missing)
	}
}

func TestRestaurantRepository_FindOwnerID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	repo := restaurant.NewRepository(integrationDB)

	ownerID, err := repo.FindOwnerID(context.Background(), restID)
	if err != nil {
		t.Fatalf("FindOwnerID() error = %v", err)
	}
	if ownerID != userID {
		t.Errorf("ownerID = %q, want %q", ownerID, userID)
	}

	emptyID, err := repo.FindOwnerID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("FindOwnerID(missing) error = %v", err)
	}
	if emptyID != "" {
		t.Errorf("expected empty string, got %q", emptyID)
	}
}

func TestRestaurantRepository_FindByUserID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	otherUserID := testhelper.CreateUser(t, integrationDB, "other@example.com")
	repo := restaurant.NewRepository(integrationDB)

	// create 2 restaurants for the owner, 1 for the other user
	r1 := &restaurant.Restaurant{UserID: userID, Name: "R1", Description: "D", Address: "A"}
	r2 := &restaurant.Restaurant{UserID: userID, Name: "R2", Description: "D", Address: "B"}
	rOther := &restaurant.Restaurant{UserID: otherUserID, Name: "Other", Description: "D", Address: "C"}
	for _, r := range []*restaurant.Restaurant{r1, r2, rOther} {
		if err := repo.Create(context.Background(), r); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	list, err := repo.FindByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if len(list) != 2 {
		t.Errorf("len = %d, want 2", len(list))
	}
	for _, r := range list {
		if r.UserID != userID {
			t.Errorf("unexpected UserID %q in result", r.UserID)
		}
	}
}
