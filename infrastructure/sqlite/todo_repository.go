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
	const ddl = `
		CREATE TABLE IF NOT EXISTS todo_tasks (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			project_name TEXT NOT NULL DEFAULT '',
			task_name    TEXT NOT NULL,
			UNIQUE(project_name, task_name)
		);`
	if _, err := db.Exec(ddl); err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}

	// 旧スキーマ（project_nameカラムなし）からの移行が必要か確認
	if needsMigration(db) {
		const migrate = `
			DROP TABLE todo_tasks;
			CREATE TABLE todo_tasks (
				id           INTEGER PRIMARY KEY AUTOINCREMENT,
				project_name TEXT NOT NULL DEFAULT '',
				task_name    TEXT NOT NULL,
				UNIQUE(project_name, task_name)
			);`
		if _, err := db.Exec(migrate); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}
	return &todoRepository{db: db}, nil
}

func needsMigration(db *sql.DB) bool {
	// テーブルのカラム一覧を取得
	rows, err := db.Query("PRAGMA table_info(todo_tasks)")
	if err != nil {
		return false
	}
	defer rows.Close()

	// 各カラムを走査し、project_nameが存在すれば移行不要
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false
		}
		if name == "project_name" {
			return false
		}
	}
	return true
}

func (r *todoRepository) FindAll(ctx context.Context) ([]*domain.TodoTask, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, project_name, task_name FROM todo_tasks ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	var todos []*domain.TodoTask
	for rows.Next() {
		t := &domain.TodoTask{}
		if err := rows.Scan(&t.ID, &t.ProjectName, &t.TaskName); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

func (r *todoRepository) Upsert(ctx context.Context, projectName, taskName string) error {
	_, err := r.db.ExecContext(ctx, "INSERT OR IGNORE INTO todo_tasks (project_name, task_name) VALUES (?, ?)", projectName, taskName)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}
	return nil
}

func (r *todoRepository) Delete(ctx context.Context, projectName, taskName string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM todo_tasks WHERE project_name = ? AND task_name = ?", projectName, taskName)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}
