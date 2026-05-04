//go:build integration

package table_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"table-service.pl/internal/modules/restaurant/table"
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

func TestTableRepository_Create_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	repo := table.NewRepository(integrationDB)

	tbl := &table.Table{RestaurantID: restID, Number: 1, Capacity: 4}
	if err := repo.Create(context.Background(), tbl); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if tbl.ID == "" {
		t.Error("ID should be set by RETURNING")
	}
}

func TestTableRepository_FindByID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := table.NewRepository(integrationDB)

	found, err := repo.FindByID(context.Background(), tblID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found == nil || found.Number != 1 || found.Capacity != 4 {
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

func TestTableRepository_FindByRestaurantID_OrderedByNumber_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	repo := table.NewRepository(integrationDB)

	// Insert in reverse order to verify sort
	for _, n := range []int{3, 1, 2} {
		tbl := &table.Table{RestaurantID: restID, Number: n, Capacity: 4}
		if err := repo.Create(context.Background(), tbl); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	list, err := repo.FindByRestaurantID(context.Background(), restID)
	if err != nil {
		t.Fatalf("FindByRestaurantID() error = %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("len = %d, want 3", len(list))
	}
	for i, want := range []int{1, 2, 3} {
		if list[i].Number != want {
			t.Errorf("list[%d].Number = %d, want %d", i, list[i].Number, want)
		}
	}
}

func TestTableRepository_FindRestaurantID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := table.NewRepository(integrationDB)

	got, err := repo.FindRestaurantID(context.Background(), tblID)
	if err != nil {
		t.Fatalf("FindRestaurantID() error = %v", err)
	}
	if got != restID {
		t.Errorf("restaurantID = %q, want %q", got, restID)
	}

	empty, err := repo.FindRestaurantID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("FindRestaurantID(missing) error = %v", err)
	}
	if empty != "" {
		t.Errorf("expected empty string, got %q", empty)
	}
}

func TestTableRepository_UpdateCapacity_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := table.NewRepository(integrationDB)

	if err := repo.UpdateCapacity(context.Background(), tblID, 8); err != nil {
		t.Fatalf("UpdateCapacity() error = %v", err)
	}

	found, _ := repo.FindByID(context.Background(), tblID)
	if found == nil || found.Capacity != 8 {
		t.Errorf("capacity after update = %+v", found)
	}

	if err := repo.UpdateCapacity(context.Background(), "00000000-0000-0000-0000-000000000000", 8); err != sql.ErrNoRows {
		t.Errorf("UpdateCapacity(missing) error = %v, want sql.ErrNoRows", err)
	}
}

func TestTableRepository_Delete_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := table.NewRepository(integrationDB)

	if err := repo.Delete(context.Background(), tblID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	found, _ := repo.FindByID(context.Background(), tblID)
	if found != nil {
		t.Error("expected table to be deleted, still found")
	}

	if err := repo.Delete(context.Background(), tblID); err != sql.ErrNoRows {
		t.Errorf("Delete(already deleted) error = %v, want sql.ErrNoRows", err)
	}
}

func TestTableRepository_UniqueNumberConstraint_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	repo := table.NewRepository(integrationDB)

	t1 := &table.Table{RestaurantID: restID, Number: 5, Capacity: 4}
	if err := repo.Create(context.Background(), t1); err != nil {
		t.Fatalf("Create first table: %v", err)
	}

	t2 := &table.Table{RestaurantID: restID, Number: 5, Capacity: 2}
	if err := repo.Create(context.Background(), t2); err == nil {
		t.Error("expected unique constraint violation, got nil")
	}
}
