package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/SkipEveryLunch/slack-toggle-syncer/domain"
	_ "modernc.org/sqlite"
)

type todoRepository struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) (domain.TodoRepository, error) {
	const ddl = `CREATE TABLE IF NOT EXISTS todo_tasks (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		task_name TEXT NOT NULL UNIQUE
	);`
	if _, err := db.Exec(ddl); err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}
	return &todoRepository{db: db}, nil
}

func (r *todoRepository) FindAll(ctx context.Context) ([]*domain.TodoTask, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, task_name FROM todo_tasks ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var todos []*domain.TodoTask
	for rows.Next() {
		t := &domain.TodoTask{}
		if err := rows.Scan(&t.ID, &t.TaskName); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

func (r *todoRepository) Upsert(ctx context.Context, taskName string) error {
	_, err := r.db.ExecContext(ctx, "INSERT OR IGNORE INTO todo_tasks (task_name) VALUES (?)", taskName)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	return nil
}

func (r *todoRepository) Delete(ctx context.Context, taskName string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM todo_tasks WHERE task_name = ?", taskName)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}
