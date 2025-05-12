package ws

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

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (db.User, error) {
	return r.queries.GetUserByUsername(ctx, username)
}
