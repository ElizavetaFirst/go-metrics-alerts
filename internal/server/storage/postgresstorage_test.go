package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ElizavetaFirst/go-metrics-alerts/internal/server/db"
	"github.com/stretchr/testify/require"
)

func setupDB(t *testing.T) *db.FakeDB {
	return &db.FakeDB{}
}

func TestUpdate(t *testing.T) {
	mockDB := setupDB(t)

	// Настройка ожидаемого exec запроса
	mockDB.expectExec()

	testStorage := PostgresStorage{db: mockDB}

	updateOpts := &UpdateOptions{
		Context:    context.Background(),
		MetricName: "test",
		Update:     Metric{Type: "test", Value: 50},
	}

	err := testStorage.Update(updateOpts)

	require.NoError(t, err)
}

func TestGetAll(t *testing.T) {
	mockDB := setupDB(t)

	// Настройка ожидаемого query запроса
	mockDB.expectQuery()

	testStorage := PostgresStorage{db: mockDB}

	getAllOpts := &GetAllOptions{
		Context: context.Background(),
	}

	metrics := testStorage.GetAll(getAllOpts)

	expectedMetric := Metric{
		Type:  "test",
		Value: "50",
	}
	require.Equal(t, metrics["test_test"], expectedMetric)

	require.NoError(t, mockDB.mock.ExpectationsWereMet())
}
