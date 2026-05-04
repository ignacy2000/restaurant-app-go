//go:build integration

package call_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"table-service.pl/internal/modules/restaurant/call"
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

func TestCallRepository_Create_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := call.NewRepository(integrationDB)

	c := &call.WaiterCall{RestaurantID: restID, TableID: tblID, Status: call.StatusPending}
	if err := repo.Create(context.Background(), c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if c.ID == "" {
		t.Error("ID should be set by RETURNING")
	}
	if c.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set by RETURNING")
	}
}

func TestCallRepository_FindByID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := call.NewRepository(integrationDB)

	c := &call.WaiterCall{RestaurantID: restID, TableID: tblID, Status: call.StatusPending}
	if err := repo.Create(context.Background(), c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.FindByID(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found == nil || found.Status != call.StatusPending {
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

func TestCallRepository_FindByRestaurantID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := call.NewRepository(integrationDB)

	for range 3 {
		c := &call.WaiterCall{RestaurantID: restID, TableID: tblID, Status: call.StatusPending}
		if err := repo.Create(context.Background(), c); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	list, err := repo.FindByRestaurantID(context.Background(), restID)
	if err != nil {
		t.Fatalf("FindByRestaurantID() error = %v", err)
	}
	if len(list) != 3 {
		t.Errorf("len = %d, want 3", len(list))
	}

	empty, err := repo.FindByRestaurantID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("FindByRestaurantID(missing) error = %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected empty, got %v", empty)
	}
}

func TestCallRepository_UpdateStatus_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := call.NewRepository(integrationDB)

	c := &call.WaiterCall{RestaurantID: restID, TableID: tblID, Status: call.StatusPending}
	if err := repo.Create(context.Background(), c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.UpdateStatus(context.Background(), c.ID, call.StatusAcknowledged); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	found, _ := repo.FindByID(context.Background(), c.ID)
	if found == nil || found.Status != call.StatusAcknowledged {
		t.Errorf("status after update: %+v", found)
	}

	if err := repo.UpdateStatus(context.Background(), "00000000-0000-0000-0000-000000000000", call.StatusDone); err != sql.ErrNoRows {
		t.Errorf("UpdateStatus(missing) error = %v, want sql.ErrNoRows", err)
	}
}

func TestCallRepository_StatusProgression_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := call.NewRepository(integrationDB)

	c := &call.WaiterCall{RestaurantID: restID, TableID: tblID, Status: call.StatusPending}
	repo.Create(context.Background(), c) //nolint:errcheck

	for _, status := range []call.CallStatus{call.StatusAcknowledged, call.StatusDone} {
		if err := repo.UpdateStatus(context.Background(), c.ID, status); err != nil {
			t.Fatalf("UpdateStatus(%s) error = %v", status, err)
		}
	}

	found, _ := repo.FindByID(context.Background(), c.ID)
	if found == nil || found.Status != call.StatusDone {
		t.Errorf("final status: %+v", found)
	}
}
