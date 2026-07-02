package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct{ db *sql.DB }

func newPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS tasks (
		id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
		title VARCHAR(120) NOT NULL,
		done BOOLEAN NOT NULL DEFAULT FALSE,
		created TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) ready(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	return s.db.PingContext(ctx)
}

func (s *PostgresStore) list() ([]Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, `SELECT id,title,done,created FROM tasks ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tasks := []Task{}
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.Done, &task.Created); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (s *PostgresStore) add(title string) (Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var task Task
	err := s.db.QueryRowContext(ctx, `INSERT INTO tasks (title) VALUES ($1) RETURNING id,title,done,created`, title).Scan(&task.ID, &task.Title, &task.Done, &task.Created)
	return task, err
}

func (s *PostgresStore) update(id int, title string, done bool) (Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var task Task
	err := s.db.QueryRowContext(ctx, `UPDATE tasks SET title=$2,done=$3 WHERE id=$1 RETURNING id,title,done,created`, id, title, done).Scan(&task.ID, &task.Title, &task.Done, &task.Created)
	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, os.ErrNotExist
	}
	return task, err
}

func (s *PostgresStore) delete(id int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id=$1`, id)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n > 0, err
}
