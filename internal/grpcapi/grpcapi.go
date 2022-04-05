package grpcapi

import (
	"github.com/Sugar-pack/users-manager/internal/logging"
	usersPb "github.com/Sugar-pack/users-manager/pkg/generated/users"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
)

type UsersService struct {
	usersPb.UnimplementedUsersServer
	dbConn *sqlx.DB
}

func CreateServer(logger logging.Logger, dbConn *sqlx.DB) (*grpc.Server, error) {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			WithLogger(logger),
			WithUniqTraceID,
			LogBoundaries,
		),
	)

	usersService := &UsersService{
		dbConn: dbConn,
	}
	usersPb.RegisterUsersServer(grpcServer, usersService)

	return grpcServer, nil
}
