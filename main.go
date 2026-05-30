package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"

	"github.com/SkipEveryLunch/slack-toggle-syncer/application"
	"github.com/SkipEveryLunch/slack-toggle-syncer/config"
	"github.com/SkipEveryLunch/slack-toggle-syncer/domain"
	infra_slack "github.com/SkipEveryLunch/slack-toggle-syncer/infrastructure/slack"
	infra_sqlite "github.com/SkipEveryLunch/slack-toggle-syncer/infrastructure/sqlite"
	infra_toggl "github.com/SkipEveryLunch/slack-toggle-syncer/infrastructure/toggl"
	_ "modernc.org/sqlite"
)

func main() {
	cfgPath := os.Getenv("GOCONF")
	if cfgPath == "" {
		cfgPath = "config/config.toml.tmpl"
	}

	cfg, err := config.New(cfgPath)
	if err != nil {
		log.Fatalf("config.New: %v", err)
	}

	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <command> (available: sync, report, todos)", os.Args[0])
	}
	cmd := os.Args[1]

	db, err := sql.Open("sqlite", cfg.Main.DBPath)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	todoRepo, err := infra_sqlite.NewTodoRepository(db)
	if err != nil {
		log.Fatalf("infra_sqlite.NewTodoRepository: %v", err)
	}

	sourceRepo := infra_slack.NewSourceRepository(cfg.Slack)

	ctx := context.Background()

	switch cmd {
	case "sync":
		togglRepo := infra_toggl.NewTogglRepository(cfg.Toggl)

		projects := make(map[domain.ProjectName]domain.ProjectID, len(cfg.Toggl.Projects))
		for k, v := range cfg.Toggl.Projects {
			projects[domain.ProjectName(k)] = domain.ProjectID(v)
		}

		usecase := &application.SyncUsecase{
			SlackRepo:   sourceRepo,
			TogglRepo:   togglRepo,
			TodoRepo:    todoRepo,
			Projects:    projects,
			WorkspaceID: cfg.Toggl.WorkspaceID,
		}
		if err := usecase.Run(ctx); err != nil {
			slog.Error("usecase.Run", "error", err)
			os.Exit(1)
		}
		slog.Info("sync completed")

	case "report":
		usecase := &application.ReportUsecase{
			SlackRepo: sourceRepo,
		}
		if err := usecase.Run(ctx); err != nil {
			slog.Error("usecase.Run", "error", err)
			os.Exit(1)
		}

	case "todos":
		usecase := &application.TodosUsecase{
			TodoRepo: todoRepo,
		}
		if err := usecase.Run(ctx); err != nil {
			slog.Error("usecase.Run", "error", err)
			os.Exit(1)
		}

	default:
		log.Fatalf("unknown command: %s (available: sync, report, todos)", cmd)
	}
}
