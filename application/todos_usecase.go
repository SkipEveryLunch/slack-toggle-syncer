package application

import (
	"context"
	"fmt"

	"github.com/SkipEveryLunch/slack-toggle-syncer/domain"
	"github.com/SkipEveryLunch/slack-toggle-syncer/internal/sliceutil"
)

type TodosUsecase struct {
	TodoRepo domain.TodoRepository
}

func (u *TodosUsecase) Run(ctx context.Context) error {
	todos, err := u.TodoRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("todoRepo.FindAll: %w", err)
	}

	// FindAllはid昇順だが、idは新→古の順で採番されているため反転して古い順にする
	reversedTodos := sliceutil.Reversed(todos)

	summary := domain.NewTaskSummary()
	for _, t := range reversedTodos {
		summary.Add(t.ProjectName, t.TaskName)
	}

	fmt.Println(summary.Render("未完了タスク"))
	return nil
}
