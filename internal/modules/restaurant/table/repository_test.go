package table_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"table-service.pl/internal/modules/restaurant/table"
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

func TestTableRepository_Create(t *testing.T) {
	db, mock := newMockDB(t)
	repo := table.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO tables`).
		WithArgs("rest-1", 3, 4).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow("tbl-1", now, now))

	tbl := &table.Table{RestaurantID: "rest-1", Number: 3, Capacity: 4}
	if err := repo.Create(context.Background(), tbl); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if tbl.ID != "tbl-1" {
		t.Errorf("ID = %q, want tbl-1", tbl.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestTableRepository_Create_DBError(t *testing.T) {
	db, mock := newMockDB(t)
	repo := table.NewRepository(db)

	mock.ExpectQuery(`INSERT INTO tables`).
		WillReturnError(errors.New("unique violation"))

	tbl := &table.Table{RestaurantID: "rest-1", Number: 1, Capacity: 2}
	if err := repo.Create(context.Background(), tbl); err == nil {
		t.Error("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestTableRepository_FindByID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := table.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM tables WHERE id`).
		WithArgs("tbl-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "restaurant_id", "number", "capacity", "created_at", "updated_at"}).
			AddRow("tbl-1", "rest-1", 3, 4, now, now))

	got, err := repo.FindByID(context.Background(), "tbl-1")
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected table, got nil")
	}
	if got.ID != "tbl-1" || got.Number != 3 || got.Capacity != 4 {
		t.Errorf("got %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestTableRepository_FindByID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := table.NewRepository(db)

	mock.ExpectQuery(`FROM tables WHERE id`).
		WithArgs("tbl-x").
		WillReturnError(sql.ErrNoRows)

	got, err := repo.FindByID(context.Background(), "tbl-x")
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

func TestTableRepository_FindRestaurantID_Found(t *testing.T) {
	db, mock := newMockDB(t)
	repo := table.NewRepository(db)

	mock.ExpectQuery(`SELECT restaurant_id FROM tables`).
		WithArgs("tbl-1").
		WillReturnRows(sqlmock.NewRows([]string{"restaurant_id"}).AddRow("rest-1"))

	restID, err := repo.FindRestaurantID(context.Background(), "tbl-1")
	if err != nil {
		t.Fatalf("FindRestaurantID() error = %v", err)
	}
	if restID != "rest-1" {
		t.Errorf("restaurantID = %q, want rest-1", restID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestTableRepository_FindRestaurantID_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := table.NewRepository(db)

	mock.ExpectQuery(`SELECT restaurant_id FROM tables`).
		WithArgs("tbl-x").
		WillReturnError(sql.ErrNoRows)

	restID, err := repo.FindRestaurantID(context.Background(), "tbl-x")
	if err != nil {
		t.Fatalf("FindRestaurantID() error = %v", err)
	}
	if restID != "" {
		t.Errorf("restaurantID = %q, want empty", restID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestTableRepository_FindByRestaurantID(t *testing.T) {
	db, mock := newMockDB(t)
	repo := table.NewRepository(db)
	now := time.Now()

	mock.ExpectQuery(`FROM tables WHERE restaurant_id`).
		WithArgs("rest-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "restaurant_id", "number", "capacity", "created_at", "updated_at"}).
			AddRow("tbl-1", "rest-1", 1, 2, now, now).
			AddRow("tbl-2", "rest-1", 2, 4, now, now))

	list, err := repo.FindByRestaurantID(context.Background(), "rest-1")
	if err != nil {
		t.Fatalf("FindByRestaurantID() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	if list[0].Number != 1 || list[1].Number != 2 {
		t.Errorf("unexpected order: %v", list)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestTableRepository_UpdateCapacity(t *testing.T) {
	db, mock := newMockDB(t)
	repo := table.NewRepository(db)

	mock.ExpectExec(`UPDATE tables SET capacity`).
		WithArgs(6, "tbl-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.UpdateCapacity(context.Background(), "tbl-1", 6); err != nil {
		t.Fatalf("UpdateCapacity() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestTableRepository_UpdateCapacity_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := table.NewRepository(db)

	mock.ExpectExec(`UPDATE tables SET capacity`).
		WithArgs(6, "tbl-x").
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := repo.UpdateCapacity(context.Background(), "tbl-x", 6); err != sql.ErrNoRows {
		t.Errorf("UpdateCapacity() error = %v, want sql.ErrNoRows", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestTableRepository_Delete(t *testing.T) {
	db, mock := newMockDB(t)
	repo := table.NewRepository(db)

	mock.ExpectExec(`DELETE FROM tables`).
		WithArgs("tbl-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), "tbl-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestTableRepository_Delete_NotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := table.NewRepository(db)

	mock.ExpectExec(`DELETE FROM tables`).
		WithArgs("tbl-x").
		WillReturnResult(sqlmock.NewResult(0, 0))

	if err := repo.Delete(context.Background(), "tbl-x"); err != sql.ErrNoRows {
		t.Errorf("Delete() error = %v, want sql.ErrNoRows", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
