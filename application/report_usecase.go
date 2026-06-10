package application

import (
	"context"
	"fmt"
	"time"

	"github.com/SkipEveryLunch/slack-toggle-syncer/domain"
	"github.com/SkipEveryLunch/slack-toggle-syncer/internal/sliceutil"
)

type ReportUsecase struct {
	SlackRepo domain.SourceRepository
}

func (u *ReportUsecase) Run(ctx context.Context) error {
	now := time.Now().In(domain.JST)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, domain.JST)

	messages, err := u.SlackRepo.FindMessages(ctx, startOfDay, now)
	if err != nil {
		return fmt.Errorf("slackRepo.FindMessages: %w", err)
	}

	// Slack APIは新しい順で返すため、古い順に並び替えてから処理する
	reversedMessages := sliceutil.Reversed(messages)

	summary := domain.NewTaskSummary()
	for _, msg := range reversedMessages {
		parent, ok := domain.ParseParentMessage(msg)
		if !ok {
			// パースできないメッセージ（感想や雑談など）はスキップ
			continue
		}
		summary.Add(parent.ProjectName, parent.TaskName)
	}

	fmt.Println(summary.Render("今日やったこと"))
	return nil
}
