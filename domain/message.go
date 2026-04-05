package domain

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var JST = time.FixedZone("Asia/Tokyo", 9*60*60)

type ProjectName string

func (p ProjectName) IsEmpty() bool { return p == "" }

type ProjectID int64

func (p ProjectID) IsUnset() bool { return p == 0 }

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

var parentMessageRegex = regexp.MustCompile(`(?m)^PJ:\s*(.*?)\s*\nタスク:\s*(.+?)\s*$`)

// ParseParentMessage SourceMessageの最初の2行をPJ:/タスク:形式としてパースする。
// 3行目以降は無視する。PJは空文字も許容（プロジェクト未指定）。
// マッチしない場合はfalseを返す。
func ParseParentMessage(msg *SourceMessage) (*ParentMessage, bool) {
	// 最初の2行だけを対象にする
	lines := strings.SplitN(msg.Text, "\n", 3)
	target := strings.Join(lines[:min(2, len(lines))], "\n")

	matches := parentMessageRegex.FindStringSubmatch(target)
	if matches == nil {
		return nil, false
	}
	return &ParentMessage{
		ProjectName: matches[1],
		TaskName:    matches[2],
		MessageID:   msg.MessageID,
	}, true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ParseSlackTimestamp(ts string) (time.Time, error) {
	unixTime, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("strconv.ParseFloat: %w", err)
	}
	return time.Unix(int64(unixTime), 0).In(JST), nil
}
