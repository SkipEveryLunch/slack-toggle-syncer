package domain

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
)

type SyncService interface {
	SyncToday(ctx context.Context) error
}

type SyncServiceImpl struct {
	SourceRepo SourceRepository
	TogglRepo  TogglRepository
	ProjectID  int64
	Logger     *zap.Logger
}

func (s *SyncServiceImpl) SyncToday(ctx context.Context) error {
	messages, err := s.SourceRepo.FindTodayMessages(ctx)
	if err != nil {
		return fmt.Errorf("SourceRepo.FindTodayMessages: %w", err)
	}
	if len(messages) == 0 {
		s.Logger.Info("no messages found today")
		return nil
	}

	// Slack APIは新しい順で返すので古い順に並び替える
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	entries := buildTimeEntries(messages, s.ProjectID)

	for _, entry := range entries {
		if err := s.TogglRepo.CreateTimeEntry(ctx, entry); err != nil {
			return fmt.Errorf("TogglRepo.CreateTimeEntry: %w", err)
		}
		s.Logger.Info("created time entry",
			zap.String("description", entry.Description),
			zap.Time("start", entry.Start),
			zap.Time("end", entry.End),
		)
	}
	return nil
}

// buildTimeEntries メッセージ一覧をタイムエントリに変換する。
// 開始時刻 = そのメッセージの投稿時刻、終了時刻 = 次のメッセージの投稿時刻。
// 最後のメッセージの終了時刻は現在時刻。
func buildTimeEntries(messages []*SourceMessage, projectID int64) []*TimeEntry {
	now := time.Now().In(JST)
	entries := make([]*TimeEntry, len(messages))
	for i, msg := range messages {
		end := now
		if i+1 < len(messages) {
			end = messages[i+1].Timestamp
		}
		entries[i] = &TimeEntry{
			Description: msg.Text,
			Start:       msg.Timestamp,
			End:         end,
			ProjectID:   projectID,
		}
	}
	return entries
}
