package grpcapi

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Sugar-pack/users-manager/internal/db"
	usersPb "github.com/Sugar-pack/users-manager/pkg/generated/users"
	"github.com/Sugar-pack/users-manager/pkg/logging"
)

const TracerNameUsersManager = "users-manager-grpcapi"

func (us *UsersService) CreateUser(ctx context.Context, newUser *usersPb.NewUser) (*usersPb.CreatedUser, error) {
	var span trace.Span
	ctx, span = otel.Tracer(TracerNameUsersManager).Start(ctx, "CreateUser")
	defer span.End()

	dbConn := us.dbConn
	logger := logging.FromContext(ctx)
	userID := uuid.New()
	createdAt := time.Now().UTC()
	txID := uuid.New()

	newDBUser := &db.User{
		ID:        userID,
		Name:      newUser.GetName(),
		CreatedAt: createdAt,
	}

	tx, err := dbConn.BeginTxx(ctx, nil) // start regular transaction
	if err != nil {
		logger.WithError(err).Error("prepare tx failed")
		return nil, status.Error(codes.Internal, "prepare tx failed")
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			if !errors.Is(rollbackErr, sql.ErrTxDone) {
				logger.WithError(err).Error("rollback tx failed")
			}
		}
	}()

	err = db.CreateUser(ctx, tx, newDBUser) // create user in regular transaction
	if err != nil {
		logger.WithError(err).Error("create users failed")
		return nil, status.Error(codes.Internal, "create user failed")
	}

	err = db.PrepareTransaction(ctx, tx, txID.String()) // start prepared transaction (2pc) in regular transaction
	if err != nil {
		logger.WithError(err).Error("start prepared tx failed")
		return nil, status.Error(codes.Internal, "start prepared tx failed")
	}
	err = tx.Commit() // commit regular transaction, yes, this is the way
	if err != nil {
		logger.WithError(err).Error("commit tx failed")
		return nil, status.Error(codes.Internal, "commit tx failed")
	}
	us.newTxIDCh <- txID.String()
	return &usersPb.CreatedUser{
		Id:   userID.String(),
		TxId: txID.String(),
	}, nil
}
