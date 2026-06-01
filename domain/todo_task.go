package domain

import "context"

type TodoTask struct {
	ID          int64
	ProjectName string
	TaskName    string
}

type TodoRepository interface {
	FindAll(ctx context.Context) ([]*TodoTask, error)
	Upsert(ctx context.Context, projectName, taskName string) error
	Delete(ctx context.Context, projectName, taskName string) error
}
