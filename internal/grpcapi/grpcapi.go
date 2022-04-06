package grpcapi

import (
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"

	distributedTxPb "github.com/Sugar-pack/users-manager/pkg/generated/distributedtx"
	usersPb "github.com/Sugar-pack/users-manager/pkg/generated/users"
	"github.com/Sugar-pack/users-manager/pkg/logging"
)

type UsersService struct {
	usersPb.UnimplementedUsersServer
	dbConn *sqlx.DB
}

type DistributedTxService struct {
	distributedTxPb.UnimplementedDistributedTxServiceServer
	dbConn *sqlx.DB
}

func CreateServer(logger logging.Logger, dbConn *sqlx.DB) (*grpc.Server, error) {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.WithLogger(logger),
			logging.WithUniqTraceID,
			logging.LogBoundaries,
		),
	)

	usersService := &UsersService{
		dbConn: dbConn,
	}
	distributedTxService := &DistributedTxService{
		dbConn: dbConn,
	}
	usersPb.RegisterUsersServer(grpcServer, usersService)
	distributedTxPb.RegisterDistributedTxServiceServer(grpcServer, distributedTxService)

	return grpcServer, nil
}
