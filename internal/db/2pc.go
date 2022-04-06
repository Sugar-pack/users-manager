package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func PrepareTransaction(ctx context.Context, dbConn sqlx.ExecerContext, txID uuid.UUID) error {
	query := fmt.Sprintf(`PREPARE TRANSACTION '%s'`, txID.String())
	_, err := dbConn.ExecContext(ctx, query)
	return err
}

func CommitPrepared(ctx context.Context, dbConn sqlx.ExecerContext, txID string) error {
	query := fmt.Sprintf(`COMMIT PREPARED '%s'`, txID)
	_, err := dbConn.ExecContext(ctx, query)
	return err
}
