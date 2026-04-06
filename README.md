# slack-toggle-syncer

Slackのtimesチャンネルの今日の投稿を読み取り、Toggl Trackにタイムエントリとして記録するCLIツール。

## 使い方

チャンネルに以下の形式で親投稿を立てて、スレッドでマーカーを返信する。

**親投稿（最初の2行がマーカー、3行目以降は無視）**

```
PJ: backend
タスク: API設計
メモや補足は3行目以降に自由に書いてOK
```

**スレッド返信（1行目がマーカー、2行目以降は無視）**

```
開始
完了
```

| マーカー | 意味                 |
| -------- | -------------------- |
| `開始`   | 作業開始             |
| `完了`   | タスク完了           |
| `中断`   | 一時中断（後で再開） |

時刻を指定した記録漏れの救済もできる（Slackの投稿時刻より優先される）：

```
開始 09:30
完了 11:00
```

プロジェクト未指定の場合は `PJ:` を空にする：

```
PJ:
タスク: 雑務
```

`./sync.sh` を実行すると、`開始`〜`完了`/`中断` の時間がTogglにタイムエントリとして登録される。

## 動作要件

- Go 1.25以上
- macOS（Keychain を使用）

## セットアップ

### 1. Slack Appの作成

1. [Slack API](https://api.slack.com/apps) にアクセスし、**Create New App** → **From scratch** でAppを作成
2. **OAuth & Permissions** → **Bot Token Scopes** に以下のスコープを追加
   - `channels:history` — パブリックチャンネルの投稿を読み取る
   - `channels:read` — パブリックチャンネルの情報を取得する
   - `groups:history` — プライベートチャンネルの投稿を読み取る
   - `groups:read` — プライベートチャンネルの情報を取得する
3. **Install to Workspace** でワークスペースにインストール
4. 表示された **Bot User OAuth Token**（`xoxb-` で始まる）を控える

### 2. ボットをチャンネルに追加

対象のtimesチャンネルで以下のコマンドを実行してボットを招待する。招待しないと `channel_not_found` エラーになる。

```
/invite @your-app-name
```

> **Note:** プライベートチャンネルの場合は Bot Token Scopes に `groups:history` と `groups:read` が必要。

### 3. Togglの設定

| 項目                 | 取得方法                                                                                                                |
| -------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `TOGGL_API_TOKEN`    | [track.toggl.com](https://track.toggl.com) → 左下アイコン → **Profile** → **Profile Settings** → 一番下の **API Token** |
| `TOGGL_WORKSPACE_ID` | [track.toggl.com](https://track.toggl.com) → **Workspaces** → **Settings** → URLの `workspaces/xxxxxxxx` の数字部分     |

### 4. クレデンシャルの登録

macOS Keychain にクレデンシャルを登録する。`sync.sh` が Keychain から自動的に読み取る。

```bash
security add-generic-password -a "$USER" -s slack-toggle-syncer.slack_bot_token -w 'xoxb-...'
security add-generic-password -a "$USER" -s slack-toggle-syncer.slack_channel_id -w 'C0123456789'
security add-generic-password -a "$USER" -s slack-toggle-syncer.toggl_api_token -w 'toggl-token-here'
security add-generic-password -a "$USER" -s slack-toggle-syncer.toggl_workspace_id -w '12345678'
```

**チャンネルIDの確認方法:** Slackでチャンネルを右クリック → リンクをコピー → URLの末尾の `C` から始まる文字列がチャンネルID

### 5. プロジェクトマッピングの設定

`config/projects.toml.example` をコピーして `config/projects.toml` を作成し、Slackの `[project_name]` とTogglのProject IDの対応を追加する。

```bash
cp config/projects.toml.example config/projects.toml
```

```toml
[projects]
backend  = 123456789
frontend = 987654321
```

**Project IDの確認方法:** [track.toggl.com](https://track.toggl.com) → **Projects** → プロジェクトを選択 → URLの末尾の数字

マッピングにないプロジェクト名はプロジェクト未指定でTogglに登録される。

## 実行

```bash
./sync.sh
```
