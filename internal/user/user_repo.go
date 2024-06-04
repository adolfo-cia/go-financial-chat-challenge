package user

import (
	"context"
	db "financial-chat-api/db/sqlc"
)

type repository struct {
	*db.Queries
}

func NewRepository(db *db.Queries) *repository {
	return &repository{Queries: db}
}

func (r *repository) Save(ctx context.Context, user User) (User, error) {
	arg := db.CreateUserParams{
		Username:       user.Username,
		HashedPassword: user.HashedPassword}

	rawUser, err := r.CreateUser(ctx, arg)
	if err != nil {
		return User{}, err
	}

	return User{
		Username:       rawUser.Username,
		HashedPassword: rawUser.HashedPassword,
		CreatedAt:      rawUser.CreatedAt.Time,
	}, nil
}

func (r *repository) GetByUsername(ctx context.Context, username string) (User, error) {
	rawUser, err := r.GetUser(ctx, username)
	if err != nil {
		return User{}, err
	}

	return User{
		Username:       rawUser.Username,
		HashedPassword: rawUser.HashedPassword,
		CreatedAt:      rawUser.CreatedAt.Time,
	}, nil
}
