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

func (r *sourceRepository) FindTodayMessages(ctx context.Context) ([]*domain.SourceMessage, error) {
	// JSTの今日00:00〜23:59:59をUNIXタイムスタンプで指定する
	now := time.Now().In(domain.JST)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, domain.JST)
	endOfDay := startOfDay.Add(24*time.Hour - time.Second)

	params := slack.GetConversationHistoryParameters{
		ChannelID: r.channelID,
		Oldest:    fmt.Sprintf("%d", startOfDay.Unix()),
		Latest:    fmt.Sprintf("%d", endOfDay.Unix()),
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
		// Bot・システムメッセージを除外する
		if msg.BotID != "" || msg.User == "USLACKBOT" || msg.SubType != "" {
			continue
		}
		ts, err := domain.ParseSlackTimestamp(msg.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("domain.ParseSlackTimestamp: %w", err)
		}
		outs = append(outs, &domain.SourceMessage{
			Text:      msg.Text,
			Timestamp: ts,
		})
	}
	return outs, nil
}
