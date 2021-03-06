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

func (dt *DistributedTxService) Commit(ctx context.Context, req *distributedTxPb.TxToCommit) (
	*distributedTxPb.TxResponse, error,
) {
	var span trace.Span
	ctx, span = otel.Tracer(TracerNameUsersManager).Start(ctx, "Commit")
	defer span.End()
	logger := logging.FromContext(ctx)
	dbConn := dt.dbConn
	txID := req.GetTxId()
	err := db.CommitPrepared(ctx, dbConn, txID)
	if err != nil {
		logger.WithError(err).Error("commit prepared transaction failed")
		return nil, status.Error(codes.Internal, "commit prepared transaction failed")
	}
	dt.cancelTxIDCh <- txID
	return &distributedTxPb.TxResponse{}, nil
}
