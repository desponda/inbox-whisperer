package data

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDBOnce      sync.Once
	testDBInstance  *DB
	testDBContainer testcontainers.Container
	testDBCleanup   func()
)

// SetupTestDB starts a Postgres container and returns a DB and cleanup func
// fileExists is a helper to check if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func SetupTestDB(t *testing.T) (*DB, func()) {
	testDBOnce.Do(func() {
		log := func(msg string, args ...interface{}) {
			fmt.Printf("[SetupTestDB] "+msg+"\n", args...)
		}
		ctx := context.Background()
		log("Starting postgres test container...")
		pgContainer, err := postgrescontainer.Run(ctx,
			"postgres:13",
			postgrescontainer.WithDatabase("testdb"),
			postgrescontainer.WithUsername("testuser"),
			postgrescontainer.WithPassword("testpass"),
			testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(90*time.Second)),
		)
		if err != nil {
			log("FAILED to start postgres container: %v", err)
			t.Fatalf("failed to start postgres container: %v", err)
		} else {
			log("Started postgres container successfully")
		}

		dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
		log("DSN: %s", dsn)
		if err != nil {
			log("FAILED to get connection string: %v", err)
			t.Fatalf("failed to get connection string: %v", err)
		}

		// Retry connecting to pgxpool up to 5 times
		var pool *pgxpool.Pool
		for i := 0; i < 5; i++ {
			log("Connecting to Postgres test DB via pgxpool (attempt %d)...", i+1)
			pool, err = pgxpool.New(ctx, dsn)
			if err == nil {
				// Try to ping
				pingErr := pool.Ping(ctx)
				if pingErr == nil {
					log("Connected to test DB via pgxpool")
					break
				} else {
					log("Ping failed: %v", pingErr)
				}
			}
			log("FAILED to connect to test db: %v", err)
			time.Sleep(2 * time.Second)
		}
		if pool == nil || err != nil {
			t.Fatalf("failed to connect to test db after retries: %v", err)
		}

		// Use golang-migrate to run all migrations
		// Allow override via MIGRATIONS_DIR env var, otherwise use new default location
		migrationDir := os.Getenv("MIGRATIONS_DIR")
		if migrationDir == "" {
			migrationDir = "/workspaces/inbox-whisperer/migrations/image"
		}
		if _, err := os.Stat(migrationDir); err != nil {
			t.Fatalf("failed to find golang-migrate migrations directory at %s: %v", migrationDir, err)
		}
		log("Applying migrations from: %s", migrationDir)
		m, err := migrate.New(
			"file://"+migrationDir,
			dsn,
		)
		if err != nil {
			t.Fatalf("failed to create migrate instance: %v", err)
		}
		err = m.Up()
		if err != nil && err.Error() != "no change" && err != migrate.ErrNoChange {
			fmt.Printf("[SetupTestDB] MIGRATION ERROR: %+v\n", err)
			t.Fatalf("failed to apply migrations: %v", err)
		}

		testDBInstance = &DB{Pool: pool}
		testDBContainer = pgContainer
		testDBCleanup = func() {
			pool.Close()
			_ = pgContainer.Terminate(ctx)
		}
	})

	cleanup := func() {
		// Truncate all tables to clean state
		if testDBInstance != nil {
			ctx := context.Background()
			db, err := sql.Open("postgres", testDBInstance.Pool.Config().ConnString())
			if err == nil {
				tables, err := db.QueryContext(ctx, `SELECT tablename FROM pg_tables WHERE schemaname = 'public'`)
				if err == nil {
					for tables.Next() {
						var table string
						tables.Scan(&table)
						db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE \"%s\" RESTART IDENTITY CASCADE", table))
					}
				}
				tables.Close()
				db.Close()
			}
		}
	}
	return testDBInstance, cleanup
}

// StopTestDB stops the shared test container (call from TestMain)
func StopTestDB() {
	if testDBCleanup != nil {
		testDBCleanup()
	}
}
