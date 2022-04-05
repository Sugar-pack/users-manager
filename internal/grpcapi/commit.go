package grpcapi

import (
	"context"

	"github.com/Sugar-pack/users-manager/internal/db"
	distributedTxPb "github.com/Sugar-pack/users-manager/pkg/generated/distributedtx"
	"github.com/Sugar-pack/users-manager/pkg/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (dt *DistributedTxService) Commit(ctx context.Context, req *distributedTxPb.TxToCommit) (*distributedTxPb.TxResponse, error) {
	logger := logging.FromContext(ctx)
	dbConn := dt.dbConn
	txID := req.GetTxId()
	err := db.CommitPrepared(ctx, dbConn, txID)
	if err != nil {
		logger.WithError(err).Error("commit prepared transaction failed")
		return nil, status.Error(codes.Internal, "commit prepared transaction failed")
	}
	return &distributedTxPb.TxResponse{}, nil
}
