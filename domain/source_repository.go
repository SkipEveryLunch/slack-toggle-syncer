package domain

import (
	"context"
	"time"
)

type SourceRepository interface {
	FindMessages(ctx context.Context, oldest, latest time.Time) ([]*SourceMessage, error)
	FindThreadReplies(ctx context.Context, messageID string) ([]*SourceMessage, error)
}
