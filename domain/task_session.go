package domain

import (
	"log/slog"
	"regexp"
	"time"
)

type TaskSession struct {
	Start time.Time
	End   time.Time
}

var timeOverrideRegex = regexp.MustCompile(`\b(\d{2}):(\d{2})\b`)

// extractTime メッセージテキストにhh:mm形式の時刻があればその日のJST時刻を返す。
// なければゼロ値を返す。
func extractTime(text string, date time.Time) time.Time {
	m := timeOverrideRegex.FindStringSubmatch(text)
	if m == nil {
		return time.Time{}
	}
	hour := int(m[1][0]-'0')*10 + int(m[1][1]-'0')
	min := int(m[2][0]-'0')*10 + int(m[2][1]-'0')
	return time.Date(date.Year(), date.Month(), date.Day(), hour, min, 0, 0, JST)
}

// resolveTime メッセージテキストにhh:mmがあればそちらを、なければSlackのタイムスタンプを使う。
func resolveTime(reply *SourceMessage) time.Time {
	if t := extractTime(reply.Text, reply.Timestamp); !t.IsZero() {
		return t
	}
	return reply.Timestamp
}

// BuildSessions スレッド返信の:wip/:done/:todoからTaskSessionを組み立てる。
// ステートマシンで処理し、:doneや:todoのない:wipは実行時刻nowで締める。
// :wip/:done/:todoの後にhh:mm形式の時刻を書くとSlackのタイムスタンプより優先される。
func BuildSessions(replies []*SourceMessage, now time.Time) []*TaskSession {
	var sessions []*TaskSession
	var currentStart time.Time // ゼロ値 = idle状態

	for _, reply := range replies {
		switch {
		case isMarker(reply.Text, ":wip"):
			t := resolveTime(reply)
			if !currentStart.IsZero() {
				// :done/:todoなしで次の:wipが来た → 現セッションをここで締める
				sessions = append(sessions, &TaskSession{Start: currentStart, End: t})
			}
			currentStart = t
		case isMarker(reply.Text, ":done"), isMarker(reply.Text, ":todo"):
			if currentStart.IsZero() {
				// :wipなしで:done/:todoが来た → 不正なので警告してスキップ
				slog.Warn("found :done/:todo without :wip, skipping", "timestamp", reply.Timestamp)
				continue
			}
			sessions = append(sessions, &TaskSession{Start: currentStart, End: resolveTime(reply)})
			currentStart = time.Time{} // idle状態に戻す
		}
		// 上記以外のテキスト（メモ等）はスキップ
	}

	// ループ終了時にまだin_progress → 実行時刻で締める
	if !currentStart.IsZero() {
		sessions = append(sessions, &TaskSession{Start: currentStart, End: now})
	}

	return sessions
}

// isMarker テキストが指定マーカーで始まるかチェックする（後ろにhh:mmが続く場合も許容）。
func isMarker(text, marker string) bool {
	if text == marker {
		return true
	}
	if len(text) > len(marker) && text[:len(marker)] == marker && text[len(marker)] == ' ' {
		return true
	}
	return false
}
