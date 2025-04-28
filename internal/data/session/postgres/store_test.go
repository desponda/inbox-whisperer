package postgres

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/auth/models"
	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // Register pgx driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testDB struct {
	pool      *pgxpool.Pool
	container testcontainers.Container
	t         *testing.T
}

func setupTestDB(t *testing.T) *testDB {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("database system is ready to accept connections"),
			wait.ForExposedPort(),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	mappedPort, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)
	hostIP, err := container.Host(ctx)
	require.NoError(t, err)

	connStr := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", hostIP, mappedPort.Port())

	poolConfig, err := pgxpool.ParseConfig(connStr)
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	require.NoError(t, err)

	return &testDB{
		pool:      pool,
		container: container,
		t:         t,
	}
}

func (tdb *testDB) cleanup() {
	ctx := context.Background()
	if tdb.pool != nil {
		tdb.pool.Close()
	}
	if err := tdb.container.Terminate(ctx); err != nil {
		tdb.t.Logf("failed to terminate container: %v", err)
	}
}

func TestNewStore(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	tests := []struct {
		name      string
		pool      *pgxpool.Pool
		tableName string
		wantErr   bool
	}{
		{
			name:      "valid config",
			pool:      tdb.pool,
			tableName: "sessions",
			wantErr:   false,
		},
		{
			name:      "nil pool",
			pool:      nil,
			tableName: "sessions",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewStore(tt.pool, tt.tableName, 24*time.Hour)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, store)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, store)
		})
	}
}

func TestStore_Integration(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	store, err := NewStore(tdb.pool, "sessions", 24*time.Hour)
	require.NoError(t, err)
	require.NotNil(t, store)

	t.Run("session lifecycle", func(t *testing.T) {
		ctx := context.Background()

		// Create session
		session, err := store.Create(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, session.ID())

		// Set values
		userID := "test-user-123"
		session.SetUserID(userID)
		session.SetValue("key1", "value1")
		session.SetValue("key2", map[string]interface{}{"nested": "value"})

		// Save session
		err = store.Save(ctx, session)
		require.NoError(t, err)

		// Retrieve and verify
		retrieved, err := store.Get(ctx, session.ID())
		require.NoError(t, err)
		assert.Equal(t, session.ID(), retrieved.ID())
		assert.Equal(t, userID, retrieved.UserID())
		assert.Equal(t, "value1", retrieved.Values()["key1"])
		assert.Equal(t, map[string]interface{}{"nested": "value"}, retrieved.Values()["key2"])

		// Delete session
		err = store.Delete(ctx, session.ID())
		require.NoError(t, err)

		// Verify deletion
		_, err = store.Get(ctx, session.ID())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "session not found")
	})

	t.Run("concurrent access", func(t *testing.T) {
		ctx := context.Background()
		var wg sync.WaitGroup
		numGoroutines := 10

		// Create and populate a session
		session, err := store.Create(ctx)
		require.NoError(t, err)
		session.SetUserID("concurrent-test-user")

		// Add some initial values
		for i := 0; i < numGoroutines; i++ {
			session.SetValue(fmt.Sprintf("key-%d", i), i)
		}
		err = store.Save(ctx, session)
		require.NoError(t, err)

		// Test concurrent reads
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				// Read session multiple times
				for j := 0; j < 3; j++ {
					s, err := store.Get(ctx, session.ID())
					require.NoError(t, err)

					// Verify all values are present
					values := s.Values()
					for k := 0; k < numGoroutines; k++ {
						key := fmt.Sprintf("key-%d", k)
						value, exists := values[key]
						assert.True(t, exists, "value for key %s should exist", key)
						if exists {
							floatVal, ok := value.(float64)
							assert.True(t, ok, "value should be a float64")
							assert.Equal(t, float64(k), floatVal, "value for key %s should be %d", key, k)
						}
					}

					time.Sleep(time.Millisecond)
				}
			}(i)
		}

		wg.Wait()

		// Test sequential writes
		for i := 0; i < numGoroutines; i++ {
			// Get latest session
			s, err := store.Get(ctx, session.ID())
			require.NoError(t, err)

			// Update one value
			key := fmt.Sprintf("key-%d", i)
			newValue := i * 2
			s.SetValue(key, newValue)
			err = store.Save(ctx, s)
			require.NoError(t, err)

			// Verify the update
			s, err = store.Get(ctx, session.ID())
			require.NoError(t, err)
			value, exists := s.Values()[key]
			assert.True(t, exists, "updated value for key %s should exist", key)
			if exists {
				floatVal, ok := value.(float64)
				assert.True(t, ok, "value should be a float64")
				assert.Equal(t, float64(newValue), floatVal, "updated value for key %s should be %d", key, newValue)
			}
		}

		// Final verification
		final, err := store.Get(ctx, session.ID())
		require.NoError(t, err)
		values := final.Values()
		assert.Equal(t, numGoroutines, len(values), "should have all values")

		// Verify final values
		for i := 0; i < numGoroutines; i++ {
			key := fmt.Sprintf("key-%d", i)
			value, exists := values[key]
			assert.True(t, exists, "value for key %s should exist", key)
			if exists {
				floatVal, ok := value.(float64)
				assert.True(t, ok, "value should be a float64")
				assert.Equal(t, float64(i*2), floatVal, "final value for key %s should be %d", key, i*2)
			}
		}
	})
}

func TestStore_Cleanup(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	shortStore, err := NewStore(tdb.pool, "sessions_short", 100*time.Millisecond)
	require.NoError(t, err)
	require.NotNil(t, shortStore)

	longStore, err := NewStore(tdb.pool, "sessions_long", 24*time.Hour)
	require.NoError(t, err)
	require.NotNil(t, longStore)

	t.Run("cleanup expired sessions", func(t *testing.T) {
		ctx := context.Background()

		// Create multiple sessions with short duration
		sessions := make([]session.Session, 3)
		for i := range sessions {
			session, err := shortStore.Create(ctx)
			require.NoError(t, err)
			session.SetUserID(fmt.Sprintf("user-%d", i))
			err = shortStore.Save(ctx, session)
			require.NoError(t, err)
			sessions[i] = session
		}

		// Create one non-expiring session with long duration
		nonExpiring, err := longStore.Create(ctx)
		require.NoError(t, err)
		nonExpiring.SetUserID("permanent-user")
		nonExpiring.SetValue("custom_expiry", true)
		// Manually set a long expiry
		nonExpiringExpiry := time.Now().Add(24 * time.Hour)
		if ms, ok := nonExpiring.(*models.Session); ok {
			ms.SetExpiresAt(nonExpiringExpiry)
		}
		err = longStore.Save(ctx, nonExpiring)
		require.NoError(t, err)

		// Verify all sessions exist
		for _, session := range sessions {
			_, err := shortStore.Get(ctx, session.ID())
			require.NoError(t, err, "session should exist before expiry")
		}

		// Wait for short sessions to expire
		time.Sleep(200 * time.Millisecond)

		// Run cleanup
		err = shortStore.Cleanup(ctx)
		require.NoError(t, err)

		// Verify expired sessions are cleaned up
		for _, session := range sessions {
			_, err := shortStore.Get(ctx, session.ID())
			assert.Error(t, err, "session should be cleaned up")
			assert.Contains(t, err.Error(), "session not found")
		}

		// Verify non-expiring session still exists
		_, err = longStore.Get(ctx, nonExpiring.ID())
		assert.NoError(t, err, "non-expiring session should still exist")

		// Verify with direct DB query
		var count int
		err = tdb.pool.QueryRow(ctx, "SELECT COUNT(*) FROM sessions_long").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "only non-expiring session should remain in database")
	})
}

func TestStore_TableCreation(t *testing.T) {
	tdb := setupTestDB(t)
	defer tdb.cleanup()

	store, err := NewStore(tdb.pool, "sessions", 24*time.Hour)
	require.NoError(t, err)
	require.NotNil(t, store)

	// Verify table structure
	type Column struct {
		ColumnName string
		DataType   string
		IsNullable string
	}

	var columns []Column
	ctx := context.Background()
	rows, err := tdb.pool.Query(ctx, `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns 
		WHERE table_name = 'sessions'
		ORDER BY ordinal_position
	`)
	require.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var col Column
		require.NoError(t, rows.Scan(&col.ColumnName, &col.DataType, &col.IsNullable))
		columns = append(columns, col)
	}

	// Verify required columns
	expectedColumns := map[string]struct {
		DataType   string
		IsNullable string
	}{
		"id": {
			DataType:   "character varying",
			IsNullable: "NO",
		},
		"user_id": {
			DataType:   "character varying",
			IsNullable: "YES",
		},
		"created_at": {
			DataType:   "timestamp with time zone",
			IsNullable: "NO",
		},
		"expires_at": {
			DataType:   "timestamp with time zone",
			IsNullable: "NO",
		},
		"values": {
			DataType:   "jsonb",
			IsNullable: "NO",
		},
	}

	assert.Equal(t, len(expectedColumns), len(columns), "sessions table should have all required columns")
	for _, col := range columns {
		expected, ok := expectedColumns[col.ColumnName]
		assert.True(t, ok, "unexpected column: %s", col.ColumnName)
		assert.Equal(t, expected.DataType, col.DataType, "wrong type for column %s", col.ColumnName)
		assert.Equal(t, expected.IsNullable, col.IsNullable, "wrong nullable setting for column %s", col.ColumnName)
	}
}
