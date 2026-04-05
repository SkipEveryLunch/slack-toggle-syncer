# slack-toggle-syncer

Slackのtimesチャンネルの今日の投稿を読み取り、Toggl Trackにタイムエントリとして記録するCLIツール。

## 使い方

Slackのtimesチャンネルに作業内容を投稿するだけでOK。

```
朝会終わり
コードレビュー開始
APIの設計中
```

`go run .` を実行すると、各投稿の時刻を開始・終了時刻としてTogglに登録される。

## 動作要件

- Go 1.25以上
- direnv

## セットアップ

### 1. Slack Appの作成

1. [Slack API](https://api.slack.com/apps) にアクセスし、**Create New App** → **From scratch** でAppを作成
2. **OAuth & Permissions** → **Bot Token Scopes** に以下のスコープを追加
   - `channels:history` — チャンネルの投稿を読み取る
   - `channels:read` — チャンネル情報を取得する
3. **Install to Workspace** でワークスペースにインストール
4. 表示された **Bot User OAuth Token**（`xoxb-` で始まる）を控える

### 2. ボットをチャンネルに追加

対象のtimesチャンネルで以下のコマンドを実行してボットを招待する。

```
/invite @your-app-name
```

### 3. Togglの設定

| 項目 | 取得方法 |
|---|---|
| `TOGGL_API_TOKEN` | [track.toggl.com](https://track.toggl.com) → 左下アイコン → **Profile** → **Profile Settings** → 一番下の **API Token** |
| `TOGGL_WORKSPACE_ID` | [track.toggl.com](https://track.toggl.com) → **Workspaces** → **Settings** → URLの `workspaces/xxxxxxxx` の数字部分 |
| `TOGGL_PROJECT_ID` | プロジェクトに紐付けない場合は `0` |

### 4. 環境変数の設定

`.env` ファイルをプロジェクトルートに作成する。

```env
LOG_LEVEL=info
SLACK_BOT_TOKEN=xoxb-xxxxxxxxxxxxxxxxx
SLACK_CHANNEL_ID=Cxxxxxxxxxx
TOGGL_API_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TOGGL_WORKSPACE_ID=12345678
TOGGL_PROJECT_ID=0
```

**チャンネルIDの確認方法:** Slackでチャンネルを右クリック → リンクをコピー → URLの末尾の `C` から始まる文字列がチャンネルID

## 実行

```bash
make run
```
