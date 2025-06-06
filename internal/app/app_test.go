package app

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestMonitorTxService_RollbackDueTimeout(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	dbConn := sqlx.NewDb(db, "sqlmock")
	defer dbConn.Close()

	txID := "test-tx"

	mock.ExpectExec("^ROLLBACK PREPARED '\\Q" + txID + "\\E'").WillReturnResult(sqlmock.NewResult(0, 0))

	newTxIDCh := make(chan string, 1)
	cancelTxIDCh := make(chan string, 1)
	svc := CreateMonitorTxService(dbConn, newTxIDCh, cancelTxIDCh, 10*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.Serve(ctx)

	newTxIDCh <- txID

	time.Sleep(20 * time.Millisecond)
	cancel()

	assert.NoError(t, mock.ExpectationsWereMet())
}
