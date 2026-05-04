//go:build integration

package menu_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"table-service.pl/internal/modules/restaurant/menu"
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

func TestMenuRepository_Create_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	repo := menu.NewRepository(integrationDB)

	m := &menu.Menu{RestaurantID: restID, Name: "Lunch Menu", Description: "Daily specials"}
	if err := repo.Create(context.Background(), m); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if m.ID == "" {
		t.Error("ID should be set by RETURNING")
	}
}

func TestMenuRepository_FindByRestaurantID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	repo := menu.NewRepository(integrationDB)

	m1 := &menu.Menu{RestaurantID: restID, Name: "Lunch", Description: "L"}
	m2 := &menu.Menu{RestaurantID: restID, Name: "Dinner", Description: "D"}
	for _, m := range []*menu.Menu{m1, m2} {
		if err := repo.Create(context.Background(), m); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	list, err := repo.FindByRestaurantID(context.Background(), restID)
	if err != nil {
		t.Fatalf("FindByRestaurantID() error = %v", err)
	}
	if len(list) != 2 {
		t.Errorf("len = %d, want 2", len(list))
	}

	empty, err := repo.FindByRestaurantID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("FindByRestaurantID(missing) error = %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected empty, got %v", empty)
	}
}

func TestMenuRepository_CreateItem_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	repo := menu.NewRepository(integrationDB)

	m := &menu.Menu{RestaurantID: restID, Name: "Lunch", Description: "D"}
	if err := repo.Create(context.Background(), m); err != nil {
		t.Fatalf("Create menu error = %v", err)
	}

	item := &menu.MenuItem{MenuID: m.ID, Name: "Pasta", Description: "Classic", Price: 12.50, Position: 1}
	if err := repo.CreateItem(context.Background(), item); err != nil {
		t.Fatalf("CreateItem() error = %v", err)
	}
	if item.ID == "" {
		t.Error("ID should be set by RETURNING")
	}
}

func TestMenuRepository_FindItemsByRestaurantID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	repo := menu.NewRepository(integrationDB)

	m := &menu.Menu{RestaurantID: restID, Name: "Lunch", Description: "D"}
	if err := repo.Create(context.Background(), m); err != nil {
		t.Fatalf("Create menu: %v", err)
	}

	items := []*menu.MenuItem{
		{MenuID: m.ID, Name: "Pasta", Description: "C", Price: 12.50, Position: 2},
		{MenuID: m.ID, Name: "Salad", Description: "G", Price: 8.00, Position: 1},
	}
	for _, item := range items {
		if err := repo.CreateItem(context.Background(), item); err != nil {
			t.Fatalf("CreateItem() error = %v", err)
		}
	}

	list, err := repo.FindItemsByRestaurantID(context.Background(), restID)
	if err != nil {
		t.Fatalf("FindItemsByRestaurantID() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	// ordered by position ASC: Salad (1) then Pasta (2)
	if list[0].Name != "Salad" || list[1].Name != "Pasta" {
		t.Errorf("unexpected order: %v", list)
	}
}

func TestMenuRepository_FindMenuRestaurantID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	repo := menu.NewRepository(integrationDB)

	m := &menu.Menu{RestaurantID: restID, Name: "Lunch", Description: "D"}
	if err := repo.Create(context.Background(), m); err != nil {
		t.Fatalf("Create menu: %v", err)
	}

	got, err := repo.FindMenuRestaurantID(context.Background(), m.ID)
	if err != nil {
		t.Fatalf("FindMenuRestaurantID() error = %v", err)
	}
	if got != restID {
		t.Errorf("restaurantID = %q, want %q", got, restID)
	}

	missing, err := repo.FindMenuRestaurantID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("FindMenuRestaurantID(missing) error = %v", err)
	}
	if missing != "" {
		t.Errorf("expected empty string, got %q", missing)
	}
}

func TestMenuRepository_FindItemRestaurantID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	repo := menu.NewRepository(integrationDB)

	m := &menu.Menu{RestaurantID: restID, Name: "Lunch", Description: "D"}
	repo.Create(context.Background(), m) //nolint:errcheck

	item := &menu.MenuItem{MenuID: m.ID, Name: "Pasta", Description: "C", Price: 10.0, Position: 1}
	repo.CreateItem(context.Background(), item) //nolint:errcheck

	got, err := repo.FindItemRestaurantID(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("FindItemRestaurantID() error = %v", err)
	}
	if got != restID {
		t.Errorf("restaurantID = %q, want %q", got, restID)
	}
}

func TestMenuRepository_DeleteItem_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	repo := menu.NewRepository(integrationDB)

	m := &menu.Menu{RestaurantID: restID, Name: "Lunch", Description: "D"}
	repo.Create(context.Background(), m) //nolint:errcheck

	item := &menu.MenuItem{MenuID: m.ID, Name: "Pasta", Description: "C", Price: 10.0, Position: 1}
	if err := repo.CreateItem(context.Background(), item); err != nil {
		t.Fatalf("CreateItem() error = %v", err)
	}

	if err := repo.DeleteItem(context.Background(), item.ID); err != nil {
		t.Fatalf("DeleteItem() error = %v", err)
	}

	items, _ := repo.FindItemsByRestaurantID(context.Background(), restID)
	if len(items) != 0 {
		t.Errorf("expected 0 items after delete, got %d", len(items))
	}
}
