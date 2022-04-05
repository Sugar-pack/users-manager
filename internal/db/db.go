package db

import (
	"context"

	"github.com/Sugar-pack/users-manager/internal/config"
	"github.com/Sugar-pack/users-manager/internal/logging"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
)

// Connect creates new db connection
func Connect(ctx context.Context, conf *config.DB) (*sqlx.DB, error) {
	logger := logging.FromContext(ctx)
	logger.WithField("conn_string", conf.ConnString).Trace("connecting to db")
	var conn, err = sqlx.ConnectContext(ctx, "pgx", conf.ConnString)
	if err != nil {
		logger.WithError(err).Error("unable to connect to database")
		return nil, err
	}
	conn.DB.SetMaxOpenConns(conf.MaxOpenConns)
	conn.DB.SetConnMaxLifetime(conf.ConnMaxLifetime)

	return conn, err
}

// Disconnect drops db connection
func Disconnect(ctx context.Context, dbConn *sqlx.DB) error {
	logger := logging.FromContext(ctx)
	logger.Trace("disconnecting db")
	return dbConn.Close()
}
