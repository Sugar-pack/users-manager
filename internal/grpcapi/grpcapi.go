package grpcapi

import (
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"

	distributedTxPb "github.com/Sugar-pack/users-manager/pkg/generated/distributedtx"
	usersPb "github.com/Sugar-pack/users-manager/pkg/generated/users"
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

func CreateServer(grpcServer *grpc.Server, dbConn *sqlx.DB,
	newTxIDCh, cancelTxIDCh chan string,
) (*grpc.Server, error) {
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

	return grpcServer, nil
}
