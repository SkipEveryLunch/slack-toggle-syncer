package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/SkipEveryLunch/slack-toggle-syncer/domain"
)

type SyncUsecase struct {
	SlackRepo domain.SourceRepository
	TogglRepo domain.TogglRepository
	Projects  map[domain.ProjectName]domain.ProjectID
}

func (u *SyncUsecase) Run(ctx context.Context) error {
	now := time.Now().In(domain.JST)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, domain.JST)

	// 1. 今日の Toggl エントリを削除
	existing, err := u.TogglRepo.FindTodayEntries(ctx)
	if err != nil {
		return fmt.Errorf("togglRepo.FindTodayEntries: %w", err)
	}
	for _, e := range existing {
		if err := u.TogglRepo.DeleteEntry(ctx, e.ID); err != nil {
			return fmt.Errorf("togglRepo.DeleteEntry (id=%d): %w", e.ID, err)
		}
	}

	// 2. Slack から取得 → タスク構築 → Toggl 登録
	messages, err := u.SlackRepo.FindMessages(ctx, startOfDay, now)
	if err != nil {
		return fmt.Errorf("slackRepo.FindMessages: %w", err)
	}

	for _, msg := range messages {
		parent, ok := domain.ParseParentMessage(msg)
		if !ok {
			continue
		}

		replies, err := u.SlackRepo.FindThreadReplies(ctx, parent.MessageID)
		if err != nil {
			return fmt.Errorf("slackRepo.FindThreadReplies (messageID=%s): %w", parent.MessageID, err)
		}

		task, err := domain.NewSlackTask(parent, replies, u.Projects, now)
		if err != nil {
			return err
		}

		for _, entry := range task.ToTogglEntries() {
			if err := u.TogglRepo.CreateTogglEntry(ctx, entry); err != nil {
				return fmt.Errorf("togglRepo.CreateTogglEntry: %w", err)
			}
			slog.Info("created time entry",
				"task", entry.Description,
				"start", entry.Start.Format("15:04"),
				"end", entry.End.Format("15:04"),
			)
		}
	}
	return nil
}
