# slack-toggle-syncer v0.1 実装計画

## 概要

Slackのtimesチャンネルの今日の投稿を読み取り、Toggl Trackにタイムエントリとして記録するCLIツール。

---

## Slackメッセージフォーマット（v0.1）

シンプルに時刻と作業内容を1行で書く形式を採用する。

```
09:00-10:00 朝会・スタンドアップ
10:00-12:00 #backend APIの設計
13:00-14:30 コードレビュー
```

- `HH:MM-HH:MM` で開始・終了時刻を指定
- スペース区切りで作業内容（説明）
- `#タグ` でTogglのタグ指定（省略可）
- プロジェクトはデフォルトで環境変数から指定（v0.1では1プロジェクト固定）

---

## アーキテクチャ

cannonicalのDDDパターンを踏襲しつつ、v0.1の最小構成で実装する。

```
slack-toggle-syncer/
├── main.go
├── go.mod
├── config/
│   ├── config.go
│   └── config.toml.tmpl
├── domain/
│   ├── time_entry.go        # Togglタイムエントリ（ドメインエンティティ）
│   ├── message.go           # Slackメッセージ（値オブジェクト）
│   ├── source_repository.go # Slack読み取りインターフェース
│   ├── toggl_repository.go  # Toggl書き込みインターフェース
│   └── sync_service.go      # 同期サービス（インターフェース＋実装）
├── application/
│   └── sync_usecase.go      # ユースケース
├── infrastructure/
│   ├── slack/
│   │   └── source_repository.go  # Slack API実装
│   └── toggl/
│       └── toggl_repository.go   # Toggl API実装
└── interface/
    └── cli/
        └── command.go       # CLIエントリポイント
```

---

## 使用ライブラリ

| ライブラリ | 用途 |
|---|---|
| `slack-go/slack` | Slack API クライアント |
| `github.com/BurntSushi/toml` | TOML設定ファイル |
| `gliderlabs/sigil` | 設定ファイルの環境変数展開 |
| `go.uber.org/zap` | 構造化ロギング |
| `go-playground/validator` | 設定値バリデーション |

Toggl APIはREST APIをnet/httpで直接呼ぶ（公式GoクライアントはないのでHTTPで実装）。

---

## 環境変数（config.toml.tmpl）

```toml
[slack]
token      = "${SLACK_BOT_TOKEN}"
channel_id = "${SLACK_CHANNEL_ID}"

[toggl]
api_token    = "${TOGGL_API_TOKEN}"
workspace_id = "${TOGGL_WORKSPACE_ID}"
project_id   = "${TOGGL_PROJECT_ID:-}"  # 省略可
```

`.env`ファイルで管理し、`source .env` してから実行する運用。

---

## 主要ファイルの実装詳細

### `domain/message.go`

```go
type SourceMessage struct {
    Text      string
    Timestamp time.Time
    UserID    string
}

// Slackのtsフォーマット（"1234567890.123456"）をtime.Timeに変換
func ParseTimestamp(ts string) (time.Time, error)
```

### `domain/time_entry.go`

```go
type TimeEntry struct {
    Description string
    Start       time.Time
    End         time.Time
    Tags        []string
    ProjectID   int64  // 0の場合はプロジェクト未指定
}

// Slackメッセージのテキストをパース（"09:00-10:00 タスク名"形式）
func ParseFromSlackText(text string, date time.Time) (*TimeEntry, error)
```

### `domain/source_repository.go`

```go
type SourceRepository interface {
    FindTodayMessages(ctx context.Context) ([]*SourceMessage, error)
}
```

### `domain/toggl_repository.go`

```go
type TogglRepository interface {
    CreateTimeEntry(ctx context.Context, entry *TimeEntry) error
}
```

### `domain/sync_service.go`

```go
type SyncService interface {
    SyncToday(ctx context.Context) error
}

type SyncServiceImpl struct {
    SourceRepo SourceRepository
    TogglRepo  TogglRepository
    ProjectID  int64
}

func (s *SyncServiceImpl) SyncToday(ctx context.Context) error {
    // 1. Slackから今日のメッセージを取得
    // 2. 各メッセージをTimeEntryにパース（パース失敗はスキップ＋警告ログ）
    // 3. TogglにCreateTimeEntry
}
```

### `infrastructure/toggl/toggl_repository.go`

Toggl API v9のREST APIを直接叩く:
- `POST https://api.track.toggl.com/api/v9/workspaces/{workspace_id}/time_entries`
- Basic認証: `{api_token}:api_token`

```go
type CreateTimeEntryRequest struct {
    Description string   `json:"description"`
    Start       string   `json:"start"`         // ISO 8601
    Stop        string   `json:"stop"`           // ISO 8601
    Duration    int64    `json:"duration"`       // 秒数
    WorkspaceID int64    `json:"workspace_id"`
    ProjectID   int64    `json:"project_id,omitempty"`
    Tags        []string `json:"tags,omitempty"`
    CreatedWith string   `json:"created_with"`   // "slack-toggle-syncer"
}
```

---

## 実装フェーズ

各フェーズ完了時に`go run .`で動作確認できることをゴールにする。

---

### フェーズ1: 基盤（動くプロジェクトの骨格を作る）

**ゴール:** `make run` でコンテナが起動し、設定を読み込んで終了できる

- [ ] `go mod init github.com/yourusername/slack-toggle-syncer`
- [ ] 依存ライブラリを`go get`
- [ ] `.gitignore` に `.env`、`config.toml` を追加
- [ ] `config/config.toml.tmpl` 作成（Slack・Togglの設定項目を定義）
- [ ] `config/config.go` 実装（sigil で環境変数展開 → TOML パース → バリデーション）
- [ ] `main.go` 骨格（設定読み込み → ログ出力して終了）
- [ ] `Dockerfile` 作成（マルチステージビルド、タイムゾーンはAsia/Tokyo）
- [ ] `Makefile` 作成（`make run` でビルド＆コンテナ実行）

**Dockerfileの方針（cannonical踏襲）:**
```dockerfile
# Build stage
FROM golang:1.24-alpine AS builder
# ... go build -o main .

# Run stage
FROM alpine:3.21
# タイムゾーン設定（JST）
# config.toml.tmplをコピー
# non-rootユーザーで実行
```

**Makefileの方針:**
```makefile
# .envを読み込んでdocker runで環境変数として渡す
run:
    docker build -t slack-toggle-syncer .
    docker run --rm --env-file .env slack-toggle-syncer
```

---

### フェーズ2: Slack読み取り（メッセージをターミナルに表示する）

**ゴール:** `go run .` でSlackの今日のメッセージが標準出力に表示される

- [ ] `domain/message.go` - `SourceMessage` 構造体、タイムスタンプパーサー
- [ ] `domain/source_repository.go` - `SourceRepository` インターフェース定義
- [ ] `infrastructure/slack/source_repository.go` - Slack API実装
  - `conversations.history` で今日分取得（JSTのoldest/latest指定）
  - ページネーション対応（`has_more` ループ）
  - Bot・システムメッセージを除外
- [ ] `main.go` を拡張 - メッセージ取得してログ出力まで繋げる

---

### フェーズ3: Toggl書き込み（Slackの内容がTogglに登録される）

**ゴール:** `go run .` でSlackメッセージがパースされTogglにタイムエントリが作成される

- [ ] `domain/time_entry.go` - `TimeEntry` 構造体、Slackテキストのパーサー（`"09:00-10:00 タスク名"` 形式）
- [ ] `domain/toggl_repository.go` - `TogglRepository` インターフェース定義
- [ ] `infrastructure/toggl/toggl_repository.go` - Toggl API v9実装
  - `POST /api/v9/workspaces/{workspace_id}/time_entries`
  - Basic認証（`{api_token}:api_token`）
- [ ] `domain/sync_service.go` - `SyncService` 実装（メッセージ取得 → パース → Toggl登録）
- [ ] `application/sync_usecase.go` - ユースケース
- [ ] `interface/cli/command.go` + `main.go` 最終形 - CLIとして整理して完成

---

## Slack APIの注意点

- `conversations.history`には`oldest`/`latest`パラメータで今日の範囲を指定
- JST 00:00:00 → UTC 前日 15:00:00 として変換して渡す
- ページネーション: `has_more`フラグでループ処理
- Bot Tokenのスコープ: `channels:history`, `channels:read`

---

## エラーハンドリング方針（v0.1）

- パースできないメッセージ（フォーマット不一致）: WARNログを出してスキップ
- Slack API失敗: fatalf（全体を止める）
- Toggl API失敗: 個別エラーをログ出力して続行

---

## v0.2以降の展望（今回はスコープ外）

- Slackメッセージフォーマットの柔軟化（自然言語→時刻推測など）
- 重複チェック（同じSlackメッセージを2回登録しない）
- Togglプロジェクト/タスクのSlackタグからのマッピング
- テストの整備（table-driven tests、mockを使った単体テスト）
- GitHub Actionsでの定時実行（cron）
