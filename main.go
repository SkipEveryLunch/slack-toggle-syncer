package main

import (
	"log"
	"os"

	"github.com/SkipEveryLunch/slack-toggle-syncer/config"
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

	logger.Info("slack-toggle-syncer started",
		zap.String("log_level", cfg.Main.LogLevel),
		zap.String("slack_channel_id", cfg.Slack.ChannelID),
		zap.Int64("toggl_workspace_id", cfg.Toggl.WorkspaceID),
	)
}
