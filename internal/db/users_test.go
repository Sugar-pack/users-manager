package db

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func newMockDB() (*sqlx.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	return sqlx.NewDb(db, "sqlmock"), mock, nil
}

func TestCreateUser_OK(t *testing.T) {
	dbConn, mock, err := newMockDB()
	assert.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, dbConn.Close())
	})

	user := &User{
		ID:        uuid.New(),
		Name:      "tester",
		CreatedAt: time.Now(),
	}

	mock.ExpectExec(`^INSERT INTO users`).
		WithArgs(user.ID, user.Name, user.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = CreateUser(context.Background(), dbConn, user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_Error(t *testing.T) {
	dbConn, mock, err := newMockDB()
	assert.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, dbConn.Close())
	})

	user := &User{
		ID:        uuid.New(),
		Name:      "tester",
		CreatedAt: time.Now(),
	}

	mock.ExpectExec(`^INSERT INTO users`).
		WithArgs(user.ID, user.Name, user.CreatedAt).
		WillReturnError(assert.AnError)

	err = CreateUser(context.Background(), dbConn, user)
	assert.Error(t, err)
}
