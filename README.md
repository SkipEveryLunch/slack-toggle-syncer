# slack-toggle-syncer

Slackのtimesチャンネルの今日の投稿を読み取り、Toggl Trackにタイムエントリとして記録するCLIツール。

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

### 3. 環境変数の設定

`.env` ファイルをプロジェクトルートに作成する。

```env
LOG_LEVEL=info
SLACK_BOT_TOKEN=xoxb-xxxxxxxxxxxxxxxxx
SLACK_CHANNEL_ID=Cxxxxxxxxxx
TOGGL_API_TOKEN=
TOGGL_WORKSPACE_ID=
TOGGL_PROJECT_ID=0
```

**チャンネルIDの確認方法:** Slackでチャンネルを右クリック → リンクをコピー → URLの末尾の `C` から始まる文字列がチャンネルID

## 実行

```bash
make run
```
