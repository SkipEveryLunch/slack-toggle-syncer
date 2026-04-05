# slack-toggle-syncer v0.2 実装計画

## v0.1からの変更点

記法をスレッド型に刷新し、より細かい粒度でタスク・プロジェクトを管理できるようにする。

---

## 新しい書式

### 基本フォーマット

チャンネルに親投稿を立てて、スレッドで `:wip` / `:done` をテキストとして返信する。
（絵文字リアクションではなくテキスト投稿。リアクション対応はv0.3以降。）

```
[backend][API設計]       ← 親投稿（プロジェクト名・タスク名）
  └ :wip                ← 作業開始（スレッド返信）
  └ :done               ← 作業終了（スレッド返信）
```

### 終了マーカーは2種類

- `:done` — タスク完了
- `:todo` — 中断（後で再開予定）

どちらもセッションの終了時刻として同様に扱う。

### 複数セッションも対応

```
[backend][API設計]
  └ :wip    (10:00)
  └ :todo   (11:00)     ← 中断
  └ :wip    (13:00)     ← 午後に再開
  └ :done   (14:30)     ← 完了
```
→ Togglに2エントリ登録: 10:00-11:00 と 13:00-14:30

### `:done`/`:todo` がないまま `:wip` の場合

作業中とみなし、記録実行時刻を終了時刻として登録する。

```
[backend][API設計]
  └ :wip    (10:00)
  └ :wip    (11:00)     ← 10:00〜11:00 のエントリを実行時刻で締めて登録
                           11:00〜 のエントリも実行時刻で締めて登録
```

---

## 実装方針

### フェーズ1: 書式パーサーとSlackスレッド読み取り

**ゴール:** `[project][task]` + スレッドの `:wip`/`:done` を読み取ってログ出力できる

- [ ] 親投稿のパース: `[project_name][task_name]` 形式を正規表現で抽出
- [ ] スレッド返信の取得: `conversations.replies` API
- [ ] `:wip`/`:done` ペアの抽出ロジック（ペアが崩れている場合はエラー）
- [ ] `domain/message.go` を拡張（`ThreadMessage`, `TaskSession` 等）
- [ ] `infrastructure/slack/source_repository.go` にスレッド取得を追加

### フェーズ2: プロジェクトマッピング

**ゴール:** `[project_name]` をTogglのプロジェクトIDに変換できる

- [ ] `config.toml.tmpl` にプロジェクトマッピングを追加
  ```toml
  [toggl.projects]
  backend  = 123456789
  frontend = 987654321
  ```
- [ ] マッピングにないプロジェクト名はWARNログを出してプロジェクト未指定で登録

### フェーズ3: 重複登録防止

**ゴール:** `make run` を複数回実行しても同じエントリが二重登録されない

- [ ] ローカルにstate fileを保持（`.sync_state.json`）
  ```json
  {
    "synced": [
      {"thread_ts": "1234567890.000100", "wip_ts": "1234567890.000200", "done_ts": "1234567890.000300"}
    ]
  }
  ```
- [ ] 実行時にstate fileを読み込み、登録済みのペアをスキップ
- [ ] 登録成功後にstate fileを更新
- [ ] state fileは `.gitignore` に追加

---

## ドメインモデルの変化

```
v0.1: SourceMessage { Text, Timestamp }
         ↓ 1:1
      TimeEntry { Description, Start, End }

v0.2: ParentMessage { ProjectName, TaskName, ThreadTS }
         ↓ 1:N（スレッド）
      TaskSession { WipTS, DoneTS }
         ↓ 1:1
      TimeEntry { Description, Start, End, ProjectID }
```

---

## レイヤー責務の整理（v0.1の反省）

v0.1では Repository のメソッドにビジネスロジックが混入していた。v0.2では以下の責務分担を徹底する。

### Repository（infrastructure層）が持つべきもの
- 外部APIを叩いて生データを取得する
- 生データをドメインオブジェクトに変換する（マッピングのみ）
- ページネーションなどAPIの都合による処理

```go
// 薄いRepositoryのイメージ
type SourceRepository interface {
    FindMessages(ctx, oldest, latest time.Time) ([]*SourceMessage, error)
    FindThreadReplies(ctx, threadTS string) ([]*SourceMessage, error)
}
```

### Domain Service が持つべきもの
- 今日の時刻範囲の計算
- Bot/システムメッセージのフィルタリング
- `[project][task]` 形式のパース
- `:wip`/`:done`/`:todo` ペアの抽出とセッション構築
- 中断・未完了セッションのエラー処理・補完ロジック

---

## v0.3以降の展望（今回スコープ外）

- 絵文字リアクションによる `:wip`/`:done` 対応（定期実行でリアクションを常時観測する方式）
- 作業中（`:wip` だけあって `:done` がまだない）エントリの扱い（実行時刻までを暫定登録など）
- Toggl APIでプロジェクト一覧を取得してマッピングを自動生成
- テストの整備
