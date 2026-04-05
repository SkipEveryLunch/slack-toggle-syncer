package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/SkipEveryLunch/slack-toggle-syncer/domain"
)

type SyncUsecase struct {
	SyncService  domain.SyncService
	TogglService domain.TogglService
	TogglRepo    domain.TogglRepository
}

func (u *SyncUsecase) Run(ctx context.Context) error {
	if err := u.TogglService.DeleteTodayEntries(ctx); err != nil {
		return fmt.Errorf("TogglService.DeleteTodayEntries: %w", err)
	}

	entries, err := u.SyncService.BuildTodayEntries(ctx)
	if err != nil {
		return fmt.Errorf("SyncService.BuildTodayEntries: %w", err)
	}

	for _, entry := range entries {
		if err := u.TogglRepo.CreateTimeEntry(ctx, entry); err != nil {
			return fmt.Errorf("TogglRepo.CreateTimeEntry: %w", err)
		}
		slog.Info("created time entry",
			"task", entry.Description,
			"start", entry.Start.Format("15:04"),
			"end", entry.End.Format("15:04"),
		)
	}
	return nil
}
