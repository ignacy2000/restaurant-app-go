package testhelper

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"table-service.pl/internal/database"
)

// NewPostgresDB starts a PostgreSQL container, runs all migrations, and returns
// a ready *sql.DB plus a cleanup function. Panics on any setup failure so it
// can be used safely in TestMain.
//
// Uses pgx directly with TLSConfig=nil to avoid SSL negotiation failures that
// occur when connecting to Docker-hosted Postgres via database/sql + pgx driver.
func NewPostgresDB(ctx context.Context) (*sql.DB, func()) {
	pgc, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
	)
	if err != nil {
		panic(fmt.Sprintf("start postgres container: %v", err))
	}

	host, err := pgc.Host(ctx)
	if err != nil {
		panic(fmt.Sprintf("get container host: %v", err))
	}
	port, err := pgc.MappedPort(ctx, "5432/tcp")
	if err != nil {
		panic(fmt.Sprintf("get container port: %v", err))
	}

	connConfig, err := pgx.ParseConfig(fmt.Sprintf(
		"postgres://testuser:testpass@%s:%s/testdb", host, port.Port(),
	))
	if err != nil {
		panic(fmt.Sprintf("parse pgx config: %v", err))
	}
	connConfig.TLSConfig = nil // disable TLS — avoids SSL negotiation issues with Docker

	// Retry connecting: the container may need a moment even after the log wait.
	var db *sql.DB
	for attempt := 1; attempt <= 10; attempt++ {
		candidate := stdlib.OpenDB(*connConfig)
		if pingErr := candidate.PingContext(ctx); pingErr == nil {
			db = candidate
			break
		}
		candidate.Close()
		if attempt == 10 {
			panic("could not connect to test postgres after 10 attempts")
		}
		time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
	}

	if err := database.MigrateUp(db); err != nil {
		panic(fmt.Sprintf("migrate up: %v", err))
	}

	return db, func() {
		db.Close()
		_ = pgc.Terminate(ctx)
	}
}

// TruncateAll removes all rows from every table and resets sequences.
// Call at the start of each integration test to ensure a clean state.
func TruncateAll(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		TRUNCATE TABLE
			waiter_calls, order_items, orders,
			menu_items, menus, tables,
			allowed_origins, sessions, restaurants, users
		RESTART IDENTITY CASCADE`)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

// CreateUser inserts a test user and returns the generated UUID.
func CreateUser(t *testing.T, db *sql.DB, email string) string {
	t.Helper()
	var id string
	if err := db.QueryRowContext(context.Background(),
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`,
		email, "testhash").Scan(&id); err != nil {
		t.Fatalf("create user %q: %v", email, err)
	}
	return id
}

// CreateRestaurant inserts a test restaurant and returns the generated UUID.
func CreateRestaurant(t *testing.T, db *sql.DB, userID string) string {
	t.Helper()
	var id string
	if err := db.QueryRowContext(context.Background(),
		`INSERT INTO restaurants (user_id, name, description, address) VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, "Test Restaurant", "Test Desc", "123 Main St").Scan(&id); err != nil {
		t.Fatalf("create restaurant: %v", err)
	}
	return id
}

// CreateTable inserts a test table and returns the generated UUID.
func CreateTable(t *testing.T, db *sql.DB, restaurantID string, number, capacity int) string {
	t.Helper()
	var id string
	if err := db.QueryRowContext(context.Background(),
		`INSERT INTO tables (restaurant_id, number, capacity) VALUES ($1, $2, $3) RETURNING id`,
		restaurantID, number, capacity).Scan(&id); err != nil {
		t.Fatalf("create table number=%d: %v", number, err)
	}
	return id
}
