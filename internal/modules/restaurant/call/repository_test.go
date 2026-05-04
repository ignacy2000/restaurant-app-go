package call_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"table-service.pl/internal/modules/restaurant/call"
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

func TestCallRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	repo := call.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO waiter_calls`).
		WithArgs("rest-1", "tbl-1", call.StatusPending).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("call-1", now, now))

	c := &call.WaiterCall{RestaurantID: "rest-1", TableID: "tbl-1", Status: call.StatusPending}
	if err := repo.Create(context.Background(), c); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if c.ID != "call-1" {
		t.Errorf("ID = %q, want call-1", c.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCallRepository_Create_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := call.NewRepository(db)

	mock.ExpectQuery(`INSERT INTO waiter_calls`).
		WillReturnError(errors.New("db error"))

	c := &call.WaiterCall{RestaurantID: "rest-1", TableID: "tbl-1", Status: call.StatusPending}
	if err := repo.Create(context.Background(), c); err == nil {
		t.Error("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCallRepository_FindByID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := call.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM waiter_calls WHERE id`).
		WithArgs("call-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "restaurant_id", "table_id", "status", "created_at", "updated_at"}).
			AddRow("call-1", "rest-1", "tbl-1", "pending", now, now))

	got, err := repo.FindByID(context.Background(), "call-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected call, got nil")
	}
	if got.ID != "call-1" || got.Status != call.StatusPending {
		t.Errorf("got %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCallRepository_FindByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := call.NewRepository(db)

	mock.ExpectQuery(`FROM waiter_calls WHERE id`).
		WithArgs("call-x").
		WillReturnError(sql.ErrNoRows)

	got, err := repo.FindByID(context.Background(), "call-x")
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

func TestCallRepository_FindByRestaurantID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := call.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM waiter_calls WHERE restaurant_id`).
		WithArgs("rest-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "restaurant_id", "table_id", "status", "created_at", "updated_at"}).
			AddRow("call-1", "rest-1", "tbl-1", "pending", now, now).
			AddRow("call-2", "rest-1", "tbl-2", "acknowledged", now, now))

	list, err := repo.FindByRestaurantID(context.Background(), "rest-1")
	if err != nil {
		t.Fatalf("FindByRestaurantID() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	if list[0].ID != "call-1" || list[1].ID != "call-2" {
		t.Errorf("unexpected IDs: %v", list)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCallRepository_FindByRestaurantID_Empty(t *testing.T) {
	db, mock := newMockDB(t)
	repo := call.NewRepository(db)

	mock.ExpectQuery(`FROM waiter_calls WHERE restaurant_id`).
		WithArgs("rest-new").
		WillReturnRows(sqlmock.NewRows([]string{"id", "restaurant_id", "table_id", "status", "created_at", "updated_at"}))

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

func TestCallRepository_UpdateStatus(t *testing.T) {
	db, mock := newMockDB(t)
	repo := call.NewRepository(db)

	mock.ExpectExec(`UPDATE waiter_calls SET status`).
		WithArgs(call.StatusAcknowledged, "call-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateStatus(context.Background(), "call-1", call.StatusAcknowledged); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestCallRepository_UpdateStatus_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := call.NewRepository(db)

	mock.ExpectExec(`UPDATE waiter_calls SET status`).
		WithArgs(call.StatusDone, "call-x").
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := repo.UpdateStatus(context.Background(), "call-x", call.StatusDone); err != sql.ErrNoRows {
		t.Errorf("UpdateStatus() error = %v, want sql.ErrNoRows", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
