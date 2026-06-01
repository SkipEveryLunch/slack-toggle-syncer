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

	projectTasks := make(map[string][]string)
	var otherTasks []string
	for _, t := range todos {
		if t.ProjectName == "" {
			otherTasks = append(otherTasks, t.TaskName)
		} else {
			projectTasks[t.ProjectName] = append(projectTasks[t.ProjectName], t.TaskName)
		}
	}

	fmt.Println("*未完了タスク*")
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
