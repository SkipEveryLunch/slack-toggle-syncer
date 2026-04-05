package domain

import "context"

type TogglEntry struct {
	ID int64
}

type TogglRepository interface {
	FindTodayEntries(ctx context.Context) ([]*TogglEntry, error)
	DeleteEntry(ctx context.Context, id int64) error
	CreateTimeEntry(ctx context.Context, entry *TimeEntry) error
}
