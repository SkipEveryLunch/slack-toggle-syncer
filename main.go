package main

import (
	"context"
	"log"
	"os"

	"github.com/SkipEveryLunch/slack-toggle-syncer/config"
	infra_slack "github.com/SkipEveryLunch/slack-toggle-syncer/infrastructure/slack"
	"go.uber.org/zap"
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

	ctx := context.Background()

	sourceRepo := infra_slack.NewSourceRepository(cfg.Slack)
	messages, err := sourceRepo.FindTodayMessages(ctx)
	if err != nil {
		logger.Fatal("sourceRepo.FindTodayMessages", zap.Error(err))
	}

	logger.Info("fetched messages", zap.Int("count", len(messages)))
	for _, msg := range messages {
		logger.Info("message",
			zap.Time("timestamp", msg.Timestamp),
			zap.String("text", msg.Text),
		)
	}
}
