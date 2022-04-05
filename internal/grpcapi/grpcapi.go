package grpcapi

import (
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"

	usersPb "github.com/Sugar-pack/users-manager/pkg/generated/users"
	"github.com/Sugar-pack/users-manager/pkg/logging"
)

type UsersService struct {
	usersPb.UnimplementedUsersServer
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
	usersPb.RegisterUsersServer(grpcServer, usersService)
	return grpcServer, nil
}
