package menu_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"table-service.pl/internal/modules/restaurant/menu"
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

func TestMenuRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	repo := menu.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO menus`).
		WithArgs("rest-1", "Lunch Menu", "Daily specials").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("menu-1", now, now))

	m := &menu.Menu{RestaurantID: "rest-1", Name: "Lunch Menu", Description: "Daily specials"}
	if err := repo.Create(context.Background(), m); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if m.ID != "menu-1" {
		t.Errorf("ID = %q, want menu-1", m.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestMenuRepository_Create_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := menu.NewRepository(db)

	mock.ExpectQuery(`INSERT INTO menus`).
		WillReturnError(errors.New("db error"))

	m := &menu.Menu{RestaurantID: "rest-1", Name: "X"}
	if err := repo.Create(context.Background(), m); err == nil {
		t.Error("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestMenuRepository_FindByRestaurantID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := menu.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM menus WHERE restaurant_id`).
		WithArgs("rest-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "restaurant_id", "name", "description", "created_at", "updated_at"}).
			AddRow("menu-1", "rest-1", "Lunch", "Daily", now, now).
			AddRow("menu-2", "rest-1", "Dinner", "Evening", now, now))

	list, err := repo.FindByRestaurantID(context.Background(), "rest-1")
	if err != nil {
		t.Fatalf("FindByRestaurantID() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	if list[0].ID != "menu-1" {
		t.Errorf("list[0].ID = %q, want menu-1", list[0].ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestMenuRepository_FindByRestaurantID_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	repo := menu.NewRepository(db)

	mock.ExpectQuery(`FROM menus WHERE restaurant_id`).
		WithArgs("rest-new").
		WillReturnRows(sqlmock.NewRows([]string{"id", "restaurant_id", "name", "description", "created_at", "updated_at"}))

	list, err := repo.FindByRestaurantID(context.Background(), "rest-new")
	if err != nil {
		t.Fatalf("FindByRestaurantID() error = %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty, got %v", list)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestMenuRepository_CreateItem(t *testing.T) {
	db, mock := newMockDB(t)
	repo := menu.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO menu_items`).
		WithArgs("menu-1", "Pasta", "Classic pasta", 12.50, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("item-1", now))

	item := &menu.MenuItem{
		MenuID:      "menu-1",
		Name:        "Pasta",
		Description: "Classic pasta",
		Price:       12.50,
		Position:    1,
	}
	if err := repo.CreateItem(context.Background(), item); err != nil {
		t.Fatalf("CreateItem() error = %v", err)
	}
	if item.ID != "item-1" {
		t.Errorf("ID = %q, want item-1", item.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestMenuRepository_FindItemsByRestaurantID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := menu.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM menu_items mi`).
		WithArgs("rest-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "menu_id", "name", "description", "price", "position", "created_at"}).
			AddRow("item-1", "menu-1", "Pasta", "Classic", 12.50, 1, now).
			AddRow("item-2", "menu-1", "Pizza", "Margherita", 10.00, 2, now))

	items, err := repo.FindItemsByRestaurantID(context.Background(), "rest-1")
	if err != nil {
		t.Fatalf("FindItemsByRestaurantID() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len = %d, want 2", len(items))
	}
	if items[0].Name != "Pasta" || items[1].Name != "Pizza" {
		t.Errorf("unexpected items: %v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestMenuRepository_FindMenuRestaurantID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := menu.NewRepository(db)

	mock.ExpectQuery(`SELECT restaurant_id FROM menus`).
		WithArgs("menu-1").
		WillReturnRows(sqlmock.NewRows([]string{"restaurant_id"}).AddRow("rest-1"))

	restID, err := repo.FindMenuRestaurantID(context.Background(), "menu-1")
	if err != nil {
		t.Fatalf("FindMenuRestaurantID() error = %v", err)
	}
	if restID != "rest-1" {
		t.Errorf("restaurantID = %q, want rest-1", restID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestMenuRepository_FindMenuRestaurantID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := menu.NewRepository(db)

	mock.ExpectQuery(`SELECT restaurant_id FROM menus`).
		WithArgs("menu-x").
		WillReturnError(sql.ErrNoRows)

	restID, err := repo.FindMenuRestaurantID(context.Background(), "menu-x")
	if err != nil {
		t.Fatalf("FindMenuRestaurantID() error = %v", err)
	}
	if restID != "" {
		t.Errorf("restaurantID = %q, want empty", restID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestMenuRepository_FindItemRestaurantID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := menu.NewRepository(db)

	mock.ExpectQuery(`SELECT m.restaurant_id`).
		WithArgs("item-1").
		WillReturnRows(sqlmock.NewRows([]string{"restaurant_id"}).AddRow("rest-1"))

	restID, err := repo.FindItemRestaurantID(context.Background(), "item-1")
	if err != nil {
		t.Fatalf("FindItemRestaurantID() error = %v", err)
	}
	if restID != "rest-1" {
		t.Errorf("restaurantID = %q, want rest-1", restID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestMenuRepository_FindItemRestaurantID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := menu.NewRepository(db)

	mock.ExpectQuery(`SELECT m.restaurant_id`).
		WithArgs("item-x").
		WillReturnError(sql.ErrNoRows)

	restID, err := repo.FindItemRestaurantID(context.Background(), "item-x")
	if err != nil {
		t.Fatalf("FindItemRestaurantID() error = %v", err)
	}
	if restID != "" {
		t.Errorf("restaurantID = %q, want empty", restID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestMenuRepository_DeleteItem(t *testing.T) {
	db, mock := newMockDB(t)
	repo := menu.NewRepository(db)

	mock.ExpectExec(`DELETE FROM menu_items`).
		WithArgs("item-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.DeleteItem(context.Background(), "item-1"); err != nil {
		t.Fatalf("DeleteItem() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
