package main

import (
	"context"
	"log"
	"os"

	"go.uber.org/zap"

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

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("zap.NewProduction: %v", err)
	}
	defer logger.Sync()

	sourceRepo := infra_slack.NewSourceRepository(cfg.Slack)
	togglRepo := infra_toggl.NewTogglRepository(cfg.Toggl)

	syncService := &domain.SyncServiceImpl{
		SourceRepo: sourceRepo,
		TogglRepo:  togglRepo,
		ProjectID:  cfg.Toggl.ProjectID,
		Logger:     logger,
	}

	usecase := &application.SyncUsecase{SyncService: syncService}

	ctx := context.Background()
	if err := usecase.Run(ctx); err != nil {
		logger.Fatal("usecase.Run", zap.Error(err))
	}

	logger.Info("sync completed")
}
