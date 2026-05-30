package domain

import (
	"fmt"
	"time"
)

type SlackTask struct {
	Description string
	ProjectID   ProjectID
	Sessions    []*TaskSession
	Done        bool
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
	sessions, done := BuildSessions(replies, now)
	return &SlackTask{
		Description: parent.TaskName,
		ProjectID:   projectID,
		Sessions:    sessions,
		Done:        done,
	}, nil
}
