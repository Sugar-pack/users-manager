package grpcapi

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	distributedtx "github.com/Sugar-pack/users-manager/pkg/generated/distributedtx"
)

func TestCommit_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	dbConn := sqlx.NewDb(db, "sqlmock")
	dt := &DistributedTxService{dbConn: dbConn, cancelTxIDCh: make(chan string, 1)}
	mock.ExpectExec("^COMMIT PREPARED").WillReturnResult(sqlmock.NewResult(0, 1))

	_, err = dt.Commit(context.Background(), &distributedtx.TxToCommit{TxId: "1"})
	if err != nil {
		t.Fatalf("Commit: %v", err)
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

func TestCommit_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	dbConn := sqlx.NewDb(db, "sqlmock")
	dt := &DistributedTxService{dbConn: dbConn, cancelTxIDCh: make(chan string, 1)}
	mock.ExpectExec("^COMMIT PREPARED").WillReturnError(assert.AnError)

	if _, err = dt.Commit(context.Background(), &distributedtx.TxToCommit{TxId: "1"}); err == nil {
		t.Fatal("expected error")
	}
}
