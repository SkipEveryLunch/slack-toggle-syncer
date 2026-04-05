package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/SkipEveryLunch/slack-toggle-syncer/application"
	"github.com/SkipEveryLunch/slack-toggle-syncer/config"
	"github.com/SkipEveryLunch/slack-toggle-syncer/domain"
	infra_slack "github.com/SkipEveryLunch/slack-toggle-syncer/infrastructure/slack"
	infra_toggl "github.com/SkipEveryLunch/slack-toggle-syncer/infrastructure/toggl"
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

	sourceRepo := infra_slack.NewSourceRepository(cfg.Slack)
	togglRepo := infra_toggl.NewTogglRepository(cfg.Toggl)

	projects := make(map[domain.ProjectName]domain.ProjectID, len(cfg.Toggl.Projects))
	for k, v := range cfg.Toggl.Projects {
		projects[domain.ProjectName(k)] = domain.ProjectID(v)
	}

	usecase := &application.SyncUsecase{
		SlackRepo: sourceRepo,
		TogglRepo: togglRepo,
		Projects:  projects,
	}

	ctx := context.Background()
	if err := usecase.Run(ctx); err != nil {
		slog.Error("usecase.Run", "error", err)
		os.Exit(1)
	}

	slog.Info("sync completed")
}
