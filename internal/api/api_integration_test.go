package api

import (
	"testing"

	"avtest/internal/store"
	"avtest/internal/store/postgres"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	dsn = "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable"
)

func TestGoMigrationsPG(t *testing.T) {
	t.Skip()

	runMigrations(t, dsn)
}

func runMigrations(t *testing.T, dsn string) {
	t.Helper()
	db := setupTestDB(t, dsn)
	t.Cleanup(func() {
		err := db.DB.Close()
		require.NoError(t, err, "db.Close()")
	})

	logger, err := zap.NewProduction()
	require.NoError(t, err)
	r := mux.NewRouter()

	testAPI := &API{logger, r, db}
	err = testAPI.db.CreateTable()
	require.NoError(t, err)

	testModerator := &store.User{
		Email:    "testuser@mail.ru",
		Password: "testpass",
		Type:     "moderator",
	}

	err = testAPI.db.CreateUser(testModerator)
	require.NoError(t, err)

	testHouse := &store.House{
		HouseNumber: 1,
		Address:     "test address",
		YearBuilt:   2021,
		Developer:   "test dev",
	}

	err = testAPI.db.CreateHouse(testHouse)
	require.NoError(t, err)

	testFlat := &store.Flat{
		HouseNumber: 1,
		FlatNumber:  1,
		Price:       100000,
		Rooms:       2,
		Status:      "created",
		Moderator:   "",
	}

	err = testAPI.db.CreateFlat(testFlat)
	require.NoError(t, err)

	testUpdateFlat := &store.Flat{
		HouseNumber: 1,
		FlatNumber:  1,
		Status:      "on moderation",
	}

	err = testAPI.db.UpdateFlat(testUpdateFlat, "token_moderator1")
	require.NoError(t, err)

	_, err = testAPI.db.GetFlatStatus(testUpdateFlat.HouseNumber, testUpdateFlat.FlatNumber)
	require.NoError(t, err)

}

func setupTestDB(t *testing.T, dsn string) *postgres.PostgresDB {
	t.Helper()
	if dsn == "" {
		t.Fatalf("database dsn is empty")
	}

	db, err := postgres.NewPostgresDB(dsn)
	require.NoError(t, err, "postgres.NewPostgresDB() error")

	err = db.DB.Ping()
	require.NoError(t, err, "db.Ping() error")

	resetPublicSchemaPG(t, db)

	return db
}

func resetPublicSchemaPG(t *testing.T, db *postgres.PostgresDB) {
	t.Helper()
	_, err := db.DB.Exec("DROP SCHEMA IF EXISTS public CASCADE;")
	require.NoError(t, err, "db.Exec(...) drop schema")

	_, err = db.DB.Exec("CREATE SCHEMA IF NOT EXISTS public;")
	require.NoError(t, err, "db.Exec(...) create schema")
}
