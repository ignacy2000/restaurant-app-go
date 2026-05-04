package order_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"table-service.pl/internal/modules/restaurant/order"
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

func TestOrderRepository_Create_NoItems(t *testing.T) {
	db, mock := newMockDB(t)
	repo := order.NewRepository(db)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO orders`).
		WithArgs("rest-1", "tbl-1", order.StatusPending, "no onions").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("ord-1", now, now))
	mock.ExpectCommit()

	o := &order.Order{
		RestaurantID: "rest-1",
		TableID:      "tbl-1",
		Status:       order.StatusPending,
		Notes:        "no onions",
	}
	if err := repo.Create(context.Background(), o); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if o.ID != "ord-1" {
		t.Errorf("ID = %q, want ord-1", o.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestOrderRepository_Create_WithItems(t *testing.T) {
	db, mock := newMockDB(t)
	repo := order.NewRepository(db)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO orders`).
		WithArgs("rest-1", "tbl-1", order.StatusPending, "").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("ord-1", now, now))
	mock.ExpectQuery(`INSERT INTO order_items`).
		WithArgs("ord-1", "Burger", 2, "extra ketchup").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("item-1"))
	mock.ExpectQuery(`INSERT INTO order_items`).
		WithArgs("ord-1", "Fries", 1, "").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("item-2"))
	mock.ExpectCommit()

	o := &order.Order{
		RestaurantID: "rest-1",
		TableID:      "tbl-1",
		Status:       order.StatusPending,
		Items: []order.OrderItem{
			{Name: "Burger", Quantity: 2, Notes: "extra ketchup"},
			{Name: "Fries", Quantity: 1, Notes: ""},
		},
	}
	if err := repo.Create(context.Background(), o); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if o.ID != "ord-1" {
		t.Errorf("ID = %q, want ord-1", o.ID)
	}
	if o.Items[0].ID != "item-1" || o.Items[1].ID != "item-2" {
		t.Errorf("item IDs not populated: %v", o.Items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestOrderRepository_Create_RollbackOnOrderError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := order.NewRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO orders`).
		WillReturnError(errors.New("db error"))
	mock.ExpectRollback()

	o := &order.Order{RestaurantID: "rest-1", TableID: "tbl-1", Status: order.StatusPending}
	if err := repo.Create(context.Background(), o); err == nil {
		t.Error("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestOrderRepository_Create_RollbackOnItemError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := order.NewRepository(db)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO orders`).
		WithArgs("rest-1", "tbl-1", order.StatusPending, "").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("ord-1", now, now))
	mock.ExpectQuery(`INSERT INTO order_items`).
		WillReturnError(errors.New("item insert failed"))
	mock.ExpectRollback()

	o := &order.Order{
		RestaurantID: "rest-1",
		TableID:      "tbl-1",
		Status:       order.StatusPending,
		Items:        []order.OrderItem{{Name: "Burger", Quantity: 1}},
	}
	if err := repo.Create(context.Background(), o); err == nil {
		t.Error("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestOrderRepository_FindByRestaurantID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := order.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM orders WHERE restaurant_id`).
		WithArgs("rest-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "restaurant_id", "table_id", "status", "notes", "created_at", "updated_at"}).
			AddRow("ord-1", "rest-1", "tbl-1", "pending", "", now, now))
	// loadItems for ord-1
	mock.ExpectQuery(`FROM order_items WHERE order_id`).
		WithArgs("ord-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "name", "quantity", "notes"}).
			AddRow("item-1", "ord-1", "Burger", 2, ""))

	orders, err := repo.FindByRestaurantID(context.Background(), "rest-1")
	if err != nil {
		t.Fatalf("FindByRestaurantID() error = %v", err)
	}
	if len(orders) != 1 {
		t.Fatalf("len = %d, want 1", len(orders))
	}
	if len(orders[0].Items) != 1 || orders[0].Items[0].Name != "Burger" {
		t.Errorf("unexpected items: %v", orders[0].Items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestOrderRepository_FindByID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := order.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM orders WHERE id`).
		WithArgs("ord-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "restaurant_id", "table_id", "status", "notes", "created_at", "updated_at"}).
			AddRow("ord-1", "rest-1", "tbl-1", "pending", "notes", now, now))
	mock.ExpectQuery(`FROM order_items WHERE order_id`).
		WithArgs("ord-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "name", "quantity", "notes"}))

	got, err := repo.FindByID(context.Background(), "ord-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected order, got nil")
	}
	if got.ID != "ord-1" {
		t.Errorf("ID = %q, want ord-1", got.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestOrderRepository_FindByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := order.NewRepository(db)

	mock.ExpectQuery(`FROM orders WHERE id`).
		WithArgs("ord-x").
		WillReturnError(sql.ErrNoRows)

	got, err := repo.FindByID(context.Background(), "ord-x")
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

func TestOrderRepository_UpdateStatus(t *testing.T) {
	db, mock := newMockDB(t)
	repo := order.NewRepository(db)

	mock.ExpectExec(`UPDATE orders SET status`).
		WithArgs(order.StatusConfirmed, "ord-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateStatus(context.Background(), "ord-1", order.StatusConfirmed); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestOrderRepository_UpdateStatus_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := order.NewRepository(db)

	mock.ExpectExec(`UPDATE orders SET status`).
		WithArgs(order.StatusConfirmed, "ord-x").
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := repo.UpdateStatus(context.Background(), "ord-x", order.StatusConfirmed); err != sql.ErrNoRows {
		t.Errorf("UpdateStatus() error = %v, want sql.ErrNoRows", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
