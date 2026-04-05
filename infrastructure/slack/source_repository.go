package slack

import (
	"context"
	"fmt"
	"time"

	"github.com/slack-go/slack"

	"github.com/SkipEveryLunch/slack-toggle-syncer/config"
	"github.com/SkipEveryLunch/slack-toggle-syncer/domain"
)

type sourceRepository struct {
	client    *slack.Client
	channelID string
}

func NewSourceRepository(cfg config.Slack) domain.SourceRepository {
	return &sourceRepository{
		client:    slack.New(cfg.Token),
		channelID: cfg.ChannelID,
	}
}

func (r *sourceRepository) FindMessages(ctx context.Context, oldest, latest time.Time) ([]*domain.SourceMessage, error) {
	params := slack.GetConversationHistoryParameters{
		ChannelID: r.channelID,
		Oldest:    fmt.Sprintf("%d", oldest.Unix()),
		Latest:    fmt.Sprintf("%d", latest.Unix()),
		Inclusive: true,
		Limit:     200,
	}

	var allMessages []slack.Message
	for {
		history, err := r.client.GetConversationHistoryContext(ctx, &params)
		if err != nil {
			return nil, fmt.Errorf("client.GetConversationHistory: %w", err)
		}
		allMessages = append(allMessages, history.Messages...)
		if !history.HasMore {
			break
		}
		params.Cursor = history.ResponseMetaData.NextCursor
	}

	outs := make([]*domain.SourceMessage, 0, len(allMessages))
	for _, msg := range allMessages {
		ts, err := domain.ParseSlackTimestamp(msg.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("domain.ParseSlackTimestamp: %w", err)
		}
		outs = append(outs, &domain.SourceMessage{
			Text:      msg.Text,
			Timestamp: ts,
			MessageID: msg.Timestamp,
		})
	}
	return outs, nil
}

func (r *sourceRepository) FindThreadReplies(ctx context.Context, messageID string) ([]*domain.SourceMessage, error) {
	replies, _, _, err := r.client.GetConversationRepliesContext(ctx, &slack.GetConversationRepliesParameters{
		ChannelID: r.channelID,
		Timestamp: messageID,
	})
	if err != nil {
		return nil, fmt.Errorf("client.GetConversationReplies: %w", err)
	}

	// 最初の要素は親投稿自体なのでスキップ
	if len(replies) <= 1 {
		return nil, nil
	}

	outs := make([]*domain.SourceMessage, 0, len(replies)-1)
	for _, msg := range replies[1:] {
		ts, err := domain.ParseSlackTimestamp(msg.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("domain.ParseSlackTimestamp: %w", err)
		}
		outs = append(outs, &domain.SourceMessage{
			Text:      msg.Text,
			Timestamp: ts,
			MessageID: msg.Timestamp,
		})
	}
	return outs, nil
}
