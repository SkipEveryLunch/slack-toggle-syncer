# slack-toggle-syncer v0.2 実装計画

## v0.1からの変更概要

- Slackの書式をスレッド型に刷新（`[project][task]` + `:wip`/`:done`/`:todo`）
- Repository を薄くしてビジネスロジックをドメイン層に移す
- プロジェクトマッピング（プロジェクト名 → Toggl Project ID）
- 重複登録防止（state file）

---

## 新しい書式

```
[backend][API設計]       ← 親投稿（プロジェクト名・タスク名）
  └ :wip                ← 作業開始（スレッド返信）
  └ :todo               ← 中断（スレッド返信）
  └ :wip                ← 再開
  └ :done               ← 完了（スレッド返信）
```

### マーカーの扱い

| パターン | 終了時刻 |
|---|---|
| `:wip` → `:done` | `:done` の投稿時刻 |
| `:wip` → `:todo` | `:todo` の投稿時刻 |
| `:wip` → `:wip` | 次の `:wip` の投稿時刻 |
| `:wip` のまま（未完了） | 実行時刻 |

---

## ドメインモデル

```
v0.1: SourceMessage { Text, Timestamp }

v0.2: SourceMessage { Text, Timestamp, MessageID } ← MessageID追加（スレッド取得に使う）
      ParentMessage { ProjectName, TaskName, MessageID }
      TaskSession   { Start, End }                  ← :wip/:done/:todo から組み立て
      TimeEntry     { Description, Start, End, ProjectID }
```

---

## フェーズ1: 書式パーサーとスレッド読み取り

**ゴール:** `go run .` で `[project][task]` + `:wip`/`:done` をパースしてログ出力できる

### domain/message.go
- `ParentMessage` 構造体を追加
  ```go
  type ParentMessage struct {
      ProjectName string
      TaskName    string
      MessageID   string // スレッド取得時のIDとして使う
  }
  ```
- `ParseParentMessage(msg SourceMessage) (*ParentMessage, bool)` を追加
  - `[project][task]` 形式を正規表現で抽出
  - マッチしない場合は `false` を返す（スキップ）

### domain/task_session.go（新規）
- `TaskSession` 構造体
  ```go
  type TaskSession struct {
      Start time.Time
      End   time.Time
  }
  ```
- `BuildSessions(replies []*SourceMessage, now time.Time) ([]*TaskSession, error)` を追加
  - `:wip`/`:done`/`:todo` のシーケンスからセッションを組み立てる
  - Bot・システムメッセージの除外もここで行う

### domain/source_repository.go
- インターフェースを変更
  ```go
  type SourceRepository interface {
      FindMessages(ctx context.Context, oldest, latest time.Time) ([]*SourceMessage, error)
      FindThreadReplies(ctx context.Context, threadTS string) ([]*SourceMessage, error)
  }
  ```

### domain/sync_service.go
- `SyncToday` のロジックを全面的に書き直す
  1. 今日の時刻範囲を計算（ここで行う）
  2. `FindMessages` で親投稿一覧を取得
  3. 各メッセージを `ParseParentMessage` でパース（マッチしないものはスキップ）
  4. 各 `ParentMessage` に対して `FindThreadReplies` でスレッド取得
  5. `BuildSessions` でセッション組み立て
  6. `TimeEntry` に変換して Toggl に登録

### infrastructure/slack/source_repository.go
- `FindMessages` に変更（時刻範囲は引数で受け取るだけ）
- `FindThreadReplies` を追加
- Bot・システムメッセージの除外を**削除**（ドメイン層に移す）

---

## フェーズ2: プロジェクトマッピング

**ゴール:** `[project_name]` から Toggl の Project ID を解決できる

### config/config.toml.tmpl
```toml
[toggl.projects]
backend  = ${TOGGL_PROJECT_BACKEND:-0}
frontend = ${TOGGL_PROJECT_FRONTEND:-0}
```

### config/config.go
- `Toggl.Projects` を `map[string]int64` で追加

### domain/sync_service.go
- `ProjectID int64` の代わりに `ProjectMap map[string]int64` を持つ
- `[project_name]` で引いてヒットしなければ `slog.Warn` を出してプロジェクト未指定（0）で登録

---

## フェーズ3: 重複登録防止

**ゴール:** `go run .` を複数回実行しても同じエントリが二重登録されない

### infrastructure/state/state_repository.go（新規）
- `.sync_state.json` をローカルファイルとして読み書き
  ```go
  type SyncedSession struct {
      ThreadTS string `json:"thread_ts"`
      WipTS    string `json:"wip_ts"`
      EndTS    string `json:"end_ts"`
  }
  ```
- `IsRegistered(threadTS, wipTS, endTS string) bool`
- `MarkRegistered(threadTS, wipTS, endTS string) error`

### domain/sync_service.go
- `StateRepository` を依存に追加
- Toggl 登録前に `IsRegistered` でチェック、登録後に `MarkRegistered`

### .gitignore
- `.sync_state.json` を追加
