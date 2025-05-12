package user

import (
	"context"
	"database/sql"
	"server/internal/db"
)

type Repository struct {
	dbPool  *sql.DB
	queries *db.Queries
}

func NewRepository(dbPool *sql.DB) Repository {
	return Repository{
		dbPool:  dbPool,
		queries: db.New(dbPool),
	}
}

func (r *Repository) CreateUser(ctx context.Context, user db.CreateUserParams) (db.User, error) {
	return r.queries.CreateUser(ctx, user)
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (db.User, error) {
	return r.queries.GetUserByUsername(ctx, username)
}

func (r *Repository) SaveRefreshToken(ctx context.Context, params db.SaveRefreshTokenParams) error {
	return r.queries.SaveRefreshToken(ctx, params)
}

func (r *Repository) IsRefreshTokenValid(ctx context.Context, params db.IsRefreshTokenValidParams) (int64, error) {
	return r.queries.IsRefreshTokenValid(ctx, params)
}

func (r *Repository) RevokeToken(ctx context.Context, jti string) error {
	return r.queries.RevokeToken(ctx, jti)
}

func (r *Repository) RevokeTokensForUser(ctx context.Context, userId string) (int64, error) {
	return r.queries.RevokeTokensForUser(ctx, userId)
}
