package domain

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type SyncService interface {
	SyncToday(ctx context.Context) error
}

type SyncServiceImpl struct {
	SourceRepo SourceRepository
	TogglRepo  TogglRepository
	ProjectID  int64
}

func (s *SyncServiceImpl) SyncToday(ctx context.Context) error {
	now := time.Now().In(JST)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, JST)

	messages, err := s.SourceRepo.FindMessages(ctx, startOfDay, now)
	if err != nil {
		return fmt.Errorf("SourceRepo.FindMessages: %w", err)
	}

	for _, msg := range messages {
		parent, ok := ParseParentMessage(msg)
		if !ok {
			continue
		}

		replies, err := s.SourceRepo.FindThreadReplies(ctx, parent.MessageID)
		if err != nil {
			return fmt.Errorf("SourceRepo.FindThreadReplies (messageID=%s): %w", parent.MessageID, err)
		}

		sessions := BuildSessions(replies, now)
		if len(sessions) == 0 {
			slog.Info("no sessions found", "project", parent.ProjectName, "task", parent.TaskName)
			continue
		}

		for _, session := range sessions {
			entry := &TimeEntry{
				Description: parent.TaskName,
				Start:       session.Start,
				End:         session.End,
				ProjectID:   s.ProjectID,
			}
			if err := s.TogglRepo.CreateTimeEntry(ctx, entry); err != nil {
				return fmt.Errorf("TogglRepo.CreateTimeEntry: %w", err)
			}
			slog.Info("created time entry",
				"project", parent.ProjectName,
				"task", parent.TaskName,
				"start", session.Start.Format("15:04"),
				"end", session.End.Format("15:04"),
			)
		}
	}
	return nil
}
