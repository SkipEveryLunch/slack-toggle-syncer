package domain

import "context"

type TodoTask struct {
	ID       int64
	TaskName string
}

type TodoRepository interface {
	FindAll(ctx context.Context) ([]*TodoTask, error)
	Upsert(ctx context.Context, taskName string) error
	Delete(ctx context.Context, taskName string) error
}
