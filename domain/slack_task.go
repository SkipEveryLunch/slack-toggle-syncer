package domain

import (
	"fmt"
	"time"
)

type SlackTask struct {
	description string
	projectID   ProjectID
	sessions    []*TaskSession
}

// NewSlackTask は ParentMessage・スレッド返信・プロジェクトマップから SlackTask を構築する。
func NewSlackTask(
	parent *ParentMessage,
	replies []*SourceMessage,
	projects map[ProjectName]ProjectID,
	now time.Time,
) (*SlackTask, error) {
	var projectID ProjectID
	if !ProjectName(parent.ProjectName).IsEmpty() {
		id, ok := projects[ProjectName(parent.ProjectName)]
		if !ok {
			return nil, fmt.Errorf("project %q is not defined in projects.toml", parent.ProjectName)
		}
		projectID = id
	}
	return &SlackTask{
		description: parent.TaskName,
		projectID:   projectID,
		sessions:    BuildSessions(replies, now),
	}, nil
}

func (t *SlackTask) ToTogglEntries() []*TogglEntry {
	entries := make([]*TogglEntry, 0, len(t.sessions))
	for _, s := range t.sessions {
		entries = append(entries, &TogglEntry{
			Description: t.description,
			Start:       s.Start,
			End:         s.End,
			ProjectID:   t.projectID,
		})
	}
	return entries
}
