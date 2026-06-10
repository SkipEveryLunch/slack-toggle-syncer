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

	// プロジェクトごとにタスク名の配列を持つMap。
	// GoのMapは順番を保証しないため、出力時の順番は不定。
	projectTasks := make(map[string][]string)
	var otherTasks []string

	// Slack APIは新しい順で返すため、古い順に並び替えてから処理する
	reversedMessages := sliceutil.Reversed(messages)

	for _, msg := range reversedMessages {
		parent, ok := domain.ParseParentMessage(msg)
		if !ok {
			// パースできないメッセージ（感想や雑談など）はスキップ
			continue
		}

		if parent.ProjectName == "" {
			otherTasks = append(otherTasks, parent.TaskName)
		} else {
			projectTasks[parent.ProjectName] = append(projectTasks[parent.ProjectName], parent.TaskName)
		}
	}

	fmt.Println("*今日やったこと*")
	for proj, tasks := range projectTasks {
		fmt.Println("*" + proj + "*")
		for _, task := range tasks {
			fmt.Println("- " + task)
		}
	}
	if len(otherTasks) > 0 {
		fmt.Println("*その他*")
		for _, task := range otherTasks {
			fmt.Println("- " + task)
		}
	}

	return nil
}
