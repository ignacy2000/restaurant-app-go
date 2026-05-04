package restaurant_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"table-service.pl/internal/modules/restaurant"
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

func TestRestaurantRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	repo := restaurant.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO restaurants`).
		WithArgs("uid-1", "La Trattoria", "Italian food", "Via Roma 1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("rest-1", now, now))

	r := &restaurant.Restaurant{
		UserID:      "uid-1",
		Name:        "La Trattoria",
		Description: "Italian food",
		Address:     "Via Roma 1",
	}
	if err := repo.Create(context.Background(), r); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if r.ID != "rest-1" {
		t.Errorf("ID = %q, want rest-1", r.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRestaurantRepository_Create_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := restaurant.NewRepository(db)

	mock.ExpectQuery(`INSERT INTO restaurants`).
		WillReturnError(errors.New("db error"))

	r := &restaurant.Restaurant{UserID: "uid-1", Name: "X"}
	if err := repo.Create(context.Background(), r); err == nil {
		t.Error("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRestaurantRepository_FindByID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := restaurant.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM restaurants WHERE id`).
		WithArgs("rest-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "description", "address", "created_at", "updated_at"}).
			AddRow("rest-1", "uid-1", "La Trattoria", "Italian food", "Via Roma 1", now, now))

	got, err := repo.FindByID(context.Background(), "rest-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected restaurant, got nil")
	}
	if got.ID != "rest-1" || got.Name != "La Trattoria" {
		t.Errorf("got %+v, want ID=rest-1 Name=La Trattoria", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRestaurantRepository_FindByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := restaurant.NewRepository(db)

	mock.ExpectQuery(`FROM restaurants WHERE id`).
		WithArgs("rest-x").
		WillReturnError(sql.ErrNoRows)

	got, err := repo.FindByID(context.Background(), "rest-x")
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

func TestRestaurantRepository_FindOwnerID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := restaurant.NewRepository(db)

	mock.ExpectQuery(`SELECT user_id FROM restaurants`).
		WithArgs("rest-1").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow("uid-1"))

	ownerID, err := repo.FindOwnerID(context.Background(), "rest-1")
	if err != nil {
		t.Fatalf("FindOwnerID() error = %v", err)
	}
	if ownerID != "uid-1" {
		t.Errorf("ownerID = %q, want uid-1", ownerID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRestaurantRepository_FindOwnerID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := restaurant.NewRepository(db)

	mock.ExpectQuery(`SELECT user_id FROM restaurants`).
		WithArgs("rest-x").
		WillReturnError(sql.ErrNoRows)

	ownerID, err := repo.FindOwnerID(context.Background(), "rest-x")
	if err != nil {
		t.Fatalf("FindOwnerID() error = %v", err)
	}
	if ownerID != "" {
		t.Errorf("ownerID = %q, want empty", ownerID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRestaurantRepository_FindByUserID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := restaurant.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM restaurants WHERE user_id`).
		WithArgs("uid-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "description", "address", "created_at", "updated_at"}).
			AddRow("rest-1", "uid-1", "La Trattoria", "Italian", "Via Roma 1", now, now).
			AddRow("rest-2", "uid-1", "Sushi Bar", "Japanese", "Tokio St 2", now, now))

	list, err := repo.FindByUserID(context.Background(), "uid-1")
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	if list[0].ID != "rest-1" {
		t.Errorf("list[0].ID = %q, want rest-1", list[0].ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRestaurantRepository_FindByUserID_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	repo := restaurant.NewRepository(db)

	mock.ExpectQuery(`FROM restaurants WHERE user_id`).
		WithArgs("uid-new").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "description", "address", "created_at", "updated_at"}))

	list, err := repo.FindByUserID(context.Background(), "uid-new")
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty, got %v", list)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
