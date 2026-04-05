package domain

import (
	"log/slog"
	"regexp"
	"strings"
	"time"
)

type TaskSession struct {
	Start time.Time
	End   time.Time
}

var timeOverrideRegex = regexp.MustCompile(`\b(\d{2}):(\d{2})\b`)

// extractTime テキストにhh:mm形式の時刻があればその日のJST時刻を返す。
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

// resolveTime 1行目にhh:mmがあればその時刻を、なければSlackのタイムスタンプを使う。
func resolveTime(reply *SourceMessage) time.Time {
	firstLine := strings.SplitN(reply.Text, "\n", 2)[0]
	if t := extractTime(firstLine, reply.Timestamp); !t.IsZero() {
		return t
	}
	return reply.Timestamp
}

// BuildSessions スレッド返信の開始/中断/完了からTaskSessionを組み立てる。
// 各返信の1行目だけをマーカーとして認識し、2行目以降は無視する。
// ステートマシンで処理し、完了/中断のない開始は実行時刻nowで締める。
// 開始/中断/完了の後にhh:mm形式の時刻を書くとSlackのタイムスタンプより優先される。
func BuildSessions(replies []*SourceMessage, now time.Time) []*TaskSession {
	var sessions []*TaskSession
	var currentStart time.Time // ゼロ値 = idle状態

	for _, reply := range replies {
		firstLine := strings.SplitN(reply.Text, "\n", 2)[0]
		switch {
		case isMarker(firstLine, "開始"):
			t := resolveTime(reply)
			if !currentStart.IsZero() {
				// 完了/中断なしで次の開始が来た → 現セッションをここで締める
				sessions = append(sessions, &TaskSession{Start: currentStart, End: t})
			}
			currentStart = t
		case isMarker(firstLine, "完了"), isMarker(firstLine, "中断"):
			if currentStart.IsZero() {
				slog.Warn("found 完了/中断 without 開始, skipping", "timestamp", reply.Timestamp)
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

// isMarker テキストの1行目が指定マーカーで始まるかチェックする（後ろにhh:mmが続く場合も許容）。
func isMarker(firstLine, marker string) bool {
	firstLine = strings.TrimSpace(firstLine)
	return firstLine == marker || strings.HasPrefix(firstLine, marker+" ")
}
