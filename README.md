# slack-toggle-syncer

Slackのtimesチャンネルの今日の投稿を読み取り、Toggl Trackにタイムエントリとして記録するCLIツール。

## 使い方

チャンネルに `[project][task]` 形式で親投稿を立てて、スレッドで `:wip` / `:done` / `:todo` を返信する。

```
[backend][API設計]       ← 親投稿
  └ :wip                ← 作業開始
  └ :done               ← 作業終了
```

| マーカー | 意味 |
|---|---|
| `:wip` | 作業開始 |
| `:done` | 完了 |
| `:todo` | 中断（後で再開） |

`go run .` を実行すると、`:wip`〜`:done`/`:todo` の時間がTogglにタイムエントリとして登録される。

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

### 4. 環境変数の設定

`.env` ファイルをプロジェクトルートに作成する。

```env
LOG_LEVEL=info
SLACK_BOT_TOKEN=xoxb-xxxxxxxxxxxxxxxxx
SLACK_CHANNEL_ID=Cxxxxxxxxxx
TOGGL_API_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
TOGGL_WORKSPACE_ID=12345678
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

### 6. direnvのセットアップ

[direnv](https://direnv.net/) を使うと、プロジェクトディレクトリに入るだけで `.env` が自動的に読み込まれる。

```bash
brew install direnv
```

シェルの設定ファイル（`~/.zshrc` など）に以下を追加する。

```bash
eval "$(direnv hook zsh)"
```

その後、プロジェクトディレクトリで一度だけ許可する。

```bash
direnv allow
```

## 実行

```bash
go run .
```
