//go:build integration

package order_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"table-service.pl/internal/modules/restaurant/order"
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

func TestOrderRepository_Create_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := order.NewRepository(integrationDB)

	o := &order.Order{
		RestaurantID: restID,
		TableID:      tblID,
		Status:       order.StatusPending,
		Notes:        "no gluten",
		Items: []order.OrderItem{
			{Name: "Burger", Quantity: 2, Notes: "well done"},
			{Name: "Fries", Quantity: 1, Notes: ""},
		},
	}
	if err := repo.Create(context.Background(), o); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if o.ID == "" {
		t.Error("order ID should be set by RETURNING")
	}
	for i, item := range o.Items {
		if item.ID == "" {
			t.Errorf("Items[%d].ID should be set by RETURNING", i)
		}
	}
}

func TestOrderRepository_FindByRestaurantID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := order.NewRepository(integrationDB)

	o1 := &order.Order{
		RestaurantID: restID,
		TableID:      tblID,
		Status:       order.StatusPending,
		Items:        []order.OrderItem{{Name: "Burger", Quantity: 1}},
	}
	o2 := &order.Order{
		RestaurantID: restID,
		TableID:      tblID,
		Status:       order.StatusConfirmed,
		Items:        []order.OrderItem{{Name: "Pizza", Quantity: 2}},
	}
	for _, o := range []*order.Order{o1, o2} {
		if err := repo.Create(context.Background(), o); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	list, err := repo.FindByRestaurantID(context.Background(), restID)
	if err != nil {
		t.Fatalf("FindByRestaurantID() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	// each order must have its items loaded
	for _, o := range list {
		if len(o.Items) == 0 {
			t.Errorf("order %s has no items loaded", o.ID)
		}
	}
}

func TestOrderRepository_FindByID_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := order.NewRepository(integrationDB)

	o := &order.Order{
		RestaurantID: restID,
		TableID:      tblID,
		Status:       order.StatusPending,
		Notes:        "extra sauce",
		Items:        []order.OrderItem{{Name: "Pasta", Quantity: 1, Notes: "al dente"}},
	}
	if err := repo.Create(context.Background(), o); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	found, err := repo.FindByID(context.Background(), o.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found == nil {
		t.Fatal("expected order, got nil")
	}
	if found.Notes != "extra sauce" {
		t.Errorf("Notes = %q, want extra sauce", found.Notes)
	}
	if len(found.Items) != 1 || found.Items[0].Name != "Pasta" {
		t.Errorf("unexpected items: %v", found.Items)
	}

	missing, err := repo.FindByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("FindByID(missing) error = %v", err)
	}
	if missing != nil {
		t.Errorf("expected nil, got %+v", missing)
	}
}

func TestOrderRepository_UpdateStatus_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := order.NewRepository(integrationDB)

	o := &order.Order{RestaurantID: restID, TableID: tblID, Status: order.StatusPending}
	if err := repo.Create(context.Background(), o); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.UpdateStatus(context.Background(), o.ID, order.StatusConfirmed); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	found, _ := repo.FindByID(context.Background(), o.ID)
	if found == nil || found.Status != order.StatusConfirmed {
		t.Errorf("status after update = %+v", found)
	}

	if err := repo.UpdateStatus(context.Background(), "00000000-0000-0000-0000-000000000000", order.StatusReady); err != sql.ErrNoRows {
		t.Errorf("UpdateStatus(missing) error = %v, want sql.ErrNoRows", err)
	}
}

func TestOrderRepository_Create_TransactionAtomicity_Integration(t *testing.T) {
	testhelper.TruncateAll(t, integrationDB)
	userID := testhelper.CreateUser(t, integrationDB, "owner@example.com")
	restID := testhelper.CreateRestaurant(t, integrationDB, userID)
	tblID := testhelper.CreateTable(t, integrationDB, restID, 1, 4)
	repo := order.NewRepository(integrationDB)

	// quantity 0 violates the CHECK (quantity > 0) constraint → causes rollback
	o := &order.Order{
		RestaurantID: restID,
		TableID:      tblID,
		Status:       order.StatusPending,
		Items:        []order.OrderItem{{Name: "Bad Item", Quantity: 0}},
	}
	if err := repo.Create(context.Background(), o); err == nil {
		t.Error("expected constraint violation, got nil")
	}

	// the order must not exist (transaction was rolled back)
	list, _ := repo.FindByRestaurantID(context.Background(), restID)
	if len(list) != 0 {
		t.Errorf("expected 0 orders after failed create, got %d", len(list))
	}
}
