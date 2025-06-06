package db

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

	"github.com/Sugar-pack/users-manager/internal/config"
)

func TestDisconnect_OK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("new mock db: %v", err)
	}
	dbConn := sqlx.NewDb(db, "sqlmock")
	mock.ExpectClose()
	if err := Disconnect(context.Background(), dbConn); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestConnect_Error(t *testing.T) {
	conf := &config.DB{ConnString: "invalid"}
	if _, err := Connect(context.Background(), conf); err == nil {
		t.Fatal("expected error")
	}
}
