package grpcapi

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	distributedtx "github.com/Sugar-pack/users-manager/pkg/generated/distributedtx"
)

func TestRollback_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	dbConn := sqlx.NewDb(db, "sqlmock")
	dt := &DistributedTxService{dbConn: dbConn, cancelTxIDCh: make(chan string, 1)}
	mock.ExpectExec("^ROLLBACK PREPARED").WillReturnResult(sqlmock.NewResult(0, 1))

	_, err = dt.Rollback(context.Background(), &distributedtx.TxToRollback{TxId: "1"})
	if err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	select {
	case txID := <-dt.cancelTxIDCh:
		assert.Equal(t, "1", txID)
	default:
		t.Fatal("no cancel tx id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestRollback_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	dbConn := sqlx.NewDb(db, "sqlmock")
	dt := &DistributedTxService{dbConn: dbConn, cancelTxIDCh: make(chan string, 1)}
	mock.ExpectExec("^ROLLBACK PREPARED").WillReturnError(assert.AnError)

	if _, err = dt.Rollback(context.Background(), &distributedtx.TxToRollback{TxId: "1"}); err == nil {
		t.Fatal("expected error")
	}
}
