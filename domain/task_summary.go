package domain

import (
	"strings"
)

// TaskSummary はプロジェクト別タスクと未分類タスクをまとめた値オブジェクト。
type TaskSummary struct {
	ProjectTasks map[string][]string
	OtherTasks   []string
}

func NewTaskSummary() *TaskSummary {
	return &TaskSummary{
		ProjectTasks: make(map[string][]string),
	}
}

// Add は projectName が空文字なら「その他」、それ以外ならプロジェクト別に振り分ける。
func (s *TaskSummary) Add(projectName, taskName string) {
	if projectName == "" {
		s.OtherTasks = append(s.OtherTasks, taskName)
		return
	}
	s.ProjectTasks[projectName] = append(s.ProjectTasks[projectName], taskName)
}

// Render は title 付きで整形済み文字列を返す。純粋関数。
func (s *TaskSummary) Render(title string) string {
	var b strings.Builder
	b.WriteString("=== " + title + " ===\n")

	// Mapの走査順は不定なため、出力順は不定（許容）
	for proj, tasks := range s.ProjectTasks {
		b.WriteString("\n[" + proj + "]\n")
		for _, t := range tasks {
			b.WriteString("  * " + t + "\n")
		}
	}
	if len(s.OtherTasks) > 0 {
		b.WriteString("\n[その他]\n")
		for _, t := range s.OtherTasks {
			b.WriteString("  * " + t + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
