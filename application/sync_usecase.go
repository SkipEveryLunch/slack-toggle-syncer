package application

import (
	"context"
	"fmt"

	"github.com/SkipEveryLunch/slack-toggle-syncer/domain"
)

type SyncUsecase struct {
	SyncService domain.SyncService
}

func (u *SyncUsecase) Run(ctx context.Context) error {
	if err := u.SyncService.SyncToday(ctx); err != nil {
		return fmt.Errorf("SyncService.SyncToday: %w", err)
	}
	return nil
}
