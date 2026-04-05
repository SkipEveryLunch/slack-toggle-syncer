package domain

import (
	"context"
	"fmt"
)

type TogglService interface {
	DeleteTodayEntries(ctx context.Context) error
}

type TogglServiceImpl struct {
	TogglRepo TogglRepository
}

func (s *TogglServiceImpl) DeleteTodayEntries(ctx context.Context) error {
	entries, err := s.TogglRepo.FindTodayEntries(ctx)
	if err != nil {
		return fmt.Errorf("TogglRepo.FindTodayEntries: %w", err)
	}
	for _, entry := range entries {
		if err := s.TogglRepo.DeleteEntry(ctx, entry.ID); err != nil {
			return fmt.Errorf("TogglRepo.DeleteEntry (id=%d): %w", entry.ID, err)
		}
	}
	return nil
}
