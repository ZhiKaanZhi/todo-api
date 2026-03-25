package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoStore struct {
	pool *pgxpool.Pool
}

func (s *TodoStore) GetAll(ctx context.Context) ([]Todo, error) {
	// query the database
}
