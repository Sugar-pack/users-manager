package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type User struct {
	CreatedAt time.Time `db:"created_at"`
	Name      string    `db:"name"`
	ID        uuid.UUID `db:"id"`
}

type NamedExecutorContext interface {
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

func CreateUser(ctx context.Context, dbConn NamedExecutorContext, user *User) error {
	query := `INSERT INTO users (id, name, created_at) VALUES (:id, :name, :created_at)`
	_, err := dbConn.NamedExecContext(ctx, query, user)
	return err
}
