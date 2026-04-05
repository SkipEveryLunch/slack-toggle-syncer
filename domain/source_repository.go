package domain

import "context"

type SourceRepository interface {
	FindTodayMessages(ctx context.Context) ([]*SourceMessage, error)
}
