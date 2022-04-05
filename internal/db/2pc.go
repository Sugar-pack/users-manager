package db

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func PrepareTransaction(ctx context.Context, dbConn sqlx.ExecerContext, txID fmt.Stringer) error {
	query := fmt.Sprintf(`PREPARE TRANSACTION '%s'`, txID.String())
	_, err := dbConn.ExecContext(ctx, query)
	return err
}

func CommitPrepared(ctx context.Context, dbConn sqlx.ExecerContext, txID string) error {
	query := fmt.Sprintf(`COMMIT PREPARED '%s'`, txID)
	_, err := dbConn.ExecContext(ctx, query)
	return err
}

func RollbackPrepared(ctx context.Context, dbConn sqlx.ExecerContext, txID string) error {
	query := fmt.Sprintf(`ROLLBACK PREPARED '%s'`, txID)
	_, err := dbConn.ExecContext(ctx, query)
	return err
}
