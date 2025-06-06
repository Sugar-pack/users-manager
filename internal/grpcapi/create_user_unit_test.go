package grpcapi

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	userspb "github.com/Sugar-pack/users-manager/pkg/generated/users"
)

func TestCreateUser_OK(t *testing.T) {
	dbSQL, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	dbConn := sqlx.NewDb(dbSQL, "sqlmock")
	us := &UsersService{dbConn: dbConn, newTxIDCh: make(chan string, 1)}

	mock.ExpectBegin()
	mock.ExpectExec("^INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("^PREPARE TRANSACTION").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	resp, err := us.CreateUser(context.Background(), &userspb.NewUser{Name: "bob"})
	if err != nil || resp == nil {
		t.Fatalf("CreateUser: %v", err)
	}
	select {
	case txID := <-us.newTxIDCh:
		assert.NotEmpty(t, txID)
	default:
		t.Fatal("no tx id")
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateUser_CreateError(t *testing.T) {
	dbSQL, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	dbConn := sqlx.NewDb(dbSQL, "sqlmock")
	us := &UsersService{dbConn: dbConn, newTxIDCh: make(chan string, 1)}

	mock.ExpectBegin()
	mock.ExpectExec("^INSERT INTO users").WillReturnError(assert.AnError)
	mock.ExpectRollback()

	_, err = us.CreateUser(context.Background(), &userspb.NewUser{Name: "bob"})
	if err == nil {
		t.Fatal("expected error")
	}
}
