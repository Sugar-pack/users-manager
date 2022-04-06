package grpcapi

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"

	"github.com/Sugar-pack/users-manager/internal/db"

	distributedTxPb "github.com/Sugar-pack/users-manager/pkg/generated/distributedtx"
	usersPb "github.com/Sugar-pack/users-manager/pkg/generated/users"
	"github.com/Sugar-pack/users-manager/pkg/logging"
)

type UsersService struct {
	usersPb.UnimplementedUsersServer
	dbConn       *sqlx.DB
	newTxIDCh    chan string
	cancelTxIDCh chan string
}

type DistributedTxService struct {
	distributedTxPb.UnimplementedDistributedTxServiceServer
	dbConn       *sqlx.DB
	newTxIDCh    chan string
	cancelTxIDCh chan string
}

func CreateServer(logger logging.Logger, dbConn *sqlx.DB) (*grpc.Server, error) {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.WithLogger(logger),
			logging.WithUniqTraceID,
			logging.LogBoundaries,
		),
	)

	newTxIDCh := make(chan string)
	cancelTxIDCh := make(chan string)
	usersService := &UsersService{
		dbConn:       dbConn,
		newTxIDCh:    newTxIDCh,
		cancelTxIDCh: cancelTxIDCh,
	}
	distributedTxService := &DistributedTxService{
		dbConn:       dbConn,
		newTxIDCh:    newTxIDCh,
		cancelTxIDCh: cancelTxIDCh,
	}
	usersPb.RegisterUsersServer(grpcServer, usersService)
	distributedTxPb.RegisterDistributedTxServiceServer(grpcServer, distributedTxService)

	ctx := context.Background()
	ctx = logging.WithContext(ctx, logger)
	go RollbackTimeouted(ctx, dbConn, newTxIDCh, cancelTxIDCh)
	return grpcServer, nil
}

func RollbackTimeouted(ctx context.Context, dbConn *sqlx.DB,
	newTxIDCh, cancelTxIDCh chan string,
) {
	logger := logging.FromContext(ctx)
	cancelation := make(map[string]*time.Timer)
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case txID := <-newTxIDCh:
			timer := time.AfterFunc(20*time.Second, func() {
				err := db.RollbackPrepared(ctx, dbConn, txID)
				if err != nil {
					logger.WithError(err).WithField("tx_id", txID).Error("rollback timeouted prepared tx failed")
				}
				logger.WithField("tx_id", txID).Trace("tx rollbacked due timeout")
			})
			logger.WithField("tx_id", txID).Trace("watching tx")
			cancelation[txID] = timer // this is not thread safe
		case txID := <-cancelTxIDCh:
			if timer, ok := cancelation[txID]; ok { // this is not thread safe
				logger.WithField("tx_id", txID).Trace("unwatching tx")
				timer.Stop()
				delete(cancelation, txID)
			}
		}
	}
}
