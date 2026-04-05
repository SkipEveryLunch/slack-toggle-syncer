package domain

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type SyncService interface {
	BuildTodayEntries(ctx context.Context) ([]*TimeEntry, error)
}

type SyncServiceImpl struct {
	SourceRepo SourceRepository
	ProjectMap map[string]int64
}

func (s *SyncServiceImpl) BuildTodayEntries(ctx context.Context) ([]*TimeEntry, error) {
	now := time.Now().In(JST)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, JST)

	messages, err := s.SourceRepo.FindMessages(ctx, startOfDay, now)
	if err != nil {
		return nil, fmt.Errorf("SourceRepo.FindMessages: %w", err)
	}

	var entries []*TimeEntry
	for _, msg := range messages {
		parent, ok := ParseParentMessage(msg)
		if !ok {
			continue
		}

		replies, err := s.SourceRepo.FindThreadReplies(ctx, parent.MessageID)
		if err != nil {
			return nil, fmt.Errorf("SourceRepo.FindThreadReplies (messageID=%s): %w", parent.MessageID, err)
		}

		sessions := BuildSessions(replies, now)
		if len(sessions) == 0 {
			slog.Info("no sessions found", "project", parent.ProjectName, "task", parent.TaskName)
			continue
		}

		var projectID int64
		if parent.ProjectName != "" {
			var ok bool
			projectID, ok = s.ProjectMap[parent.ProjectName]
			if !ok {
				return nil, fmt.Errorf("project %q is not defined in projects.toml", parent.ProjectName)
			}
		}

		for _, session := range sessions {
			entries = append(entries, &TimeEntry{
				Description: parent.TaskName,
				Start:       session.Start,
				End:         session.End,
				ProjectID:   projectID,
			})
		}
	}
	return entries, nil
}
