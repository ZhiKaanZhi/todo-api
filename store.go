package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoStorer interface {
	GetAll(ctx context.Context) ([]Todo, error)
	Create(ctx context.Context, t *Todo) error
	Delete(ctx context.Context, id int) error
}

type TodoStore struct {
	pool *pgxpool.Pool
}

func (s *TodoStore) GetAll(ctx context.Context) ([]Todo, error) {
	rows, err := s.pool.Query(ctx,
		"SELECT id, title, description, done, created_at, updated_at, expires_at FROM todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Done,
			&t.CreatedAt, &t.UpdatedAt, &t.ExpiresAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

func (s *TodoStore) Create(ctx context.Context, t *Todo) error {
	return s.pool.QueryRow(ctx,
		`INSERT INTO todos (title, description)
		 VALUES ($1, $2)
		 RETURNING id, created_at, expires_at`,
		t.Title, t.Description,
	).Scan(&t.ID, &t.CreatedAt, &t.ExpiresAt)
}

func (s *TodoStore) Delete(ctx context.Context, id int) error {
	result, err := s.pool.Exec(ctx,
		"DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("todo not found")
	}
	return nil
}
