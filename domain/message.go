package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

type SourceMessage struct {
	Text      string
	Timestamp time.Time
	MessageID string // SlackのメッセージID。親投稿のMessageIDはスレッドIDと一致する
}

type ParentMessage struct {
	ProjectName string
	TaskName    string
	MessageID   string
}

var parentMessageRegex = regexp.MustCompile(`^\[(.*?)\]\[(.+?)\]`)

// ParseParentMessage SourceMessageを[project][task]形式としてパースする。
// projectは空文字も許容する（[][task]でプロジェクト未指定）。
// マッチしない場合はfalseを返す。
func ParseParentMessage(msg *SourceMessage) (*ParentMessage, bool) {
	matches := parentMessageRegex.FindStringSubmatch(msg.Text)
	if matches == nil {
		return nil, false
	}
	return &ParentMessage{
		ProjectName: matches[1],
		TaskName:    matches[2],
		MessageID:   msg.MessageID,
	}, true
}

func ParseSlackTimestamp(ts string) (time.Time, error) {
	unixTime, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("strconv.ParseFloat: %w", err)
	}
	return time.Unix(int64(unixTime), 0).In(JST), nil
}
