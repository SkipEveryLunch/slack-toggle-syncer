package domain

import (
	"log/slog"
	"time"
)

type TaskSession struct {
	Start time.Time
	End   time.Time
}

// BuildSessions スレッド返信の:wip/:done/:todoからTaskSessionを組み立てる。
// ステートマシンで処理し、:doneや:todoのない:wipは実行時刻nowで締める。
func BuildSessions(replies []*SourceMessage, now time.Time) []*TaskSession {
	var sessions []*TaskSession
	var currentStart time.Time // ゼロ値 = idle状態

	for _, reply := range replies {
		switch reply.Text {
		case ":wip":
			if !currentStart.IsZero() {
				// :done/:todoなしで次の:wipが来た → 現セッションをここで締める
				sessions = append(sessions, &TaskSession{Start: currentStart, End: reply.Timestamp})
			}
			currentStart = reply.Timestamp
		case ":done", ":todo":
			if currentStart.IsZero() {
				// :wipなしで:done/:todoが来た → 不正なので警告してスキップ
				slog.Warn("found :done/:todo without :wip, skipping", "timestamp", reply.Timestamp)
				continue
			}
			sessions = append(sessions, &TaskSession{Start: currentStart, End: reply.Timestamp})
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
