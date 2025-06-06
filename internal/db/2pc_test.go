package db

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func newMock() (*sqlx.DB, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	return sqlx.NewDb(db, "sqlmock"), mock, nil
}

func TestPrepareTransaction_OK(t *testing.T) {
	dbConn, mock, err := newMock()
	assert.NoError(t, err)
	defer dbConn.Close()

	mock.ExpectExec("PREPARE TRANSACTION").WillReturnResult(sqlmock.NewResult(0, 1))

	err = PrepareTransaction(context.Background(), dbConn, "txid")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPrepareTransaction_Error(t *testing.T) {
	dbConn, mock, err := newMock()
	assert.NoError(t, err)
	defer dbConn.Close()

	mock.ExpectExec("PREPARE TRANSACTION").WillReturnError(assert.AnError)

	err = PrepareTransaction(context.Background(), dbConn, "txid")
	assert.Error(t, err)
}

func TestCommitPrepared_Error(t *testing.T) {
	dbConn, mock, err := newMock()
	assert.NoError(t, err)
	defer dbConn.Close()

	mock.ExpectExec("COMMIT PREPARED").WillReturnError(assert.AnError)

	err = CommitPrepared(context.Background(), dbConn, "txid")
	assert.Error(t, err)
}

func TestRollbackPrepared_OK(t *testing.T) {
	dbConn, mock, err := newMock()
	assert.NoError(t, err)
	defer dbConn.Close()

	mock.ExpectExec("ROLLBACK PREPARED").WillReturnResult(sqlmock.NewResult(0, 1))

	err = RollbackPrepared(context.Background(), dbConn, "txid")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
