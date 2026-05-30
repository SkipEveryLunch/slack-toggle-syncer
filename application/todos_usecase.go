package application

import (
	"context"
	"fmt"

	"github.com/SkipEveryLunch/slack-toggle-syncer/domain"
)

type TodosUsecase struct {
	TodoRepo domain.TodoRepository
}

func (u *TodosUsecase) Run(ctx context.Context) error {
	todos, err := u.TodoRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("todoRepo.FindAll: %w", err)
	}
	fmt.Println("*未完了タスク*")
	for _, t := range todos {
		fmt.Println("- " + t.TaskName)
	}
	return nil
}
