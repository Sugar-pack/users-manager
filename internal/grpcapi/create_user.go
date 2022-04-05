package grpcapi

import (
	"context"

	usersPb "github.com/Sugar-pack/users-manager/pkg/generated/users"
	"github.com/google/uuid"
)

func (us *UsersService) CreateUser(ctx context.Context, newUser *usersPb.NewUser) (*usersPb.CreatedUser, error) {
	userID := uuid.New()
	txID := uuid.New()

	return &usersPb.CreatedUser{
		Id:   userID.String(),
		TxId: txID.String(),
	}, nil
}
