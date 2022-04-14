//nolint:dupl // need to refactor handler
package grpcapi

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Sugar-pack/users-manager/internal/db"
	distributedTxPb "github.com/Sugar-pack/users-manager/pkg/generated/distributedtx"
	"github.com/Sugar-pack/users-manager/pkg/logging"
)

func (dt *DistributedTxService) Rollback(ctx context.Context, req *distributedTxPb.TxToRollback) (
	*distributedTxPb.TxResponse, error,
) {
	var span trace.Span
	ctx, span = otel.Tracer(TracerNameUsersManager).Start(ctx, "Rollback")
	defer span.End()
	logger := logging.FromContext(ctx)
	dbConn := dt.dbConn
	txID := req.GetTxId()
	err := db.RollbackPrepared(ctx, dbConn, txID)
	if err != nil {
		logger.WithError(err).Error("rollback prepared transaction failed")
		return nil, status.Error(codes.Internal, "rollback prepared transaction failed")
	}
	dt.cancelTxIDCh <- txID
	return &distributedTxPb.TxResponse{}, nil
}
