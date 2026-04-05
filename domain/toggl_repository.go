package domain

import "context"

type TogglRepository interface {
	CreateTimeEntry(ctx context.Context, entry *TimeEntry) error
}
