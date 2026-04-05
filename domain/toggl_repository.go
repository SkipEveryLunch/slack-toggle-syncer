package domain

import "context"

type DeleteTogglEntry struct {
	ID int64
}

type TogglRepository interface {
	FindTodayEntries(ctx context.Context) ([]*DeleteTogglEntry, error)
	DeleteEntry(ctx context.Context, id int64) error
	CreateTogglEntry(ctx context.Context, entry *TogglEntry) error
}
