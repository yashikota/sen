# sen 設計

sen は一人で使うローカルファーストな課題管理である。
Linear の密度とキーボード操作を借りるが、チーム製品の複製ではない。
データは Markdown、TOML、JSONL を正とし、SQLite は持たない。
エージェントがファイルを直接編集でき、次の API 読み込みで UI に載る。

## 誰のための sen か

利用者は常に一人である。
担当者、メンバー、Inbox、通知、権限、リアルタイム共同編集は持たない。
複数人で同じワークスペースを共有する前提も持たない。

複数マシン間の受け渡しは、自分用のスナップショットとして `sen push` と `sen pull` で行う。
マージはしない。
後勝ちの置き換えである。

## 残すもの

Issue、Board、Project、Cycle、Label、Page、キーボード操作、コマンドパレットである。
Cycle は自分の時間枠である。チームのスプリントではない。
サブ Issue は大きい仕事を自分で分解するための親子関係である。
カスタム View は、よく使う絞り込みを名前付きで残すものである。
コメントは会話ではなく、自分へのメモである。

## 含めないもの

次は人と組織のための機能であり、今後も入れない。

- 担当者、メンバー、チーム、権限
- Inbox、通知、リアルタイム同期
- ブロック関係、見積もり
- 添付、Initiative
- GitHub Issue 連携
- SQLite キャッシュ

## ワークスペース

既定はカレントディレクトリの `.sen/` である。
環境変数 `SEN_HOME` があればそれを使う。
初期化の判定は `workspace.toml` の有無である。

```
$SEN_HOME/
  workspace.toml
  labels.toml
  projects/<slug>.toml
  cycles/<n>.toml
  views/<slug>.toml
  issues/SEN-n.md
  pages/<slug>.md
  activities.jsonl
```

Issue と Page は `+++` で囲んだ TOML frontmatter と Markdown 本文である。
メモは Issue ファイルの `[[comments]]` に置く。
活動履歴は追記の JSONL である。
サブ Issue は frontmatter の `parent`（親の識別子 `SEN-n`）で表す。
循環参照は拒否する。

## ランタイム

`sen` は単一の Go バイナリである。

- `sen init`：`.sen/` と空の `workspace.toml` を作る
- `sen serve`：JSON API と SPA を `127.0.0.1:7730` で出す
- `sen push`：自分の GHCR 参照へスナップショットを送る
- `sen pull`：スナップショットでローカルを置き換える
- `sen status`：未 push の有無と最後の digest
- `sen check`：ファイルの意味的な壊れを一覧する

認証トークンはファイルに保存しない。
`GITHUB_TOKEN`、なければ `gh auth token` を使う。

開発時は Vite+ が API へプロキシする。
本番バイナリは SPA を `go:embed` で同梱する。

## データ模型

時刻は RFC3339 の UTC で保存する。
表示だけがワークスペースのタイムゾーンに従う。
API の読み書きは毎回ディスクから読み直す。

### Workspace

名前、GHCR 参照、タイムゾーン、`lastPushedAt`、`lastPushedDigest` を持つ。
ローカルに1つだけである。

### Project

名前、slug、説明、状態、開始日、目標日を持つ。
進捗は紐づく Issue の完了割合から API が算出する。

### Cycle

番号、開始、終了、状態を持つ。
状態は `upcoming`、`active`、`completed` のいずれかである。
`active` は同時に1つだけとする。
ある Cycle を `active` にしたとき、それまで `active` だった Cycle は `completed` にする。
これは自分の集中期間であり、チームのイテレーション計画ではない。

### Label

名前と色を持つ。
Issue と多対多で結ぶ。

### View

名前、slug、表示（`list` または `board`）、任意の絞り込みを持つ。
絞り込みは状態、Project、Cycle、Label、優先度である。
他人向けの共有 View や、メンバー単位のフィルタは持たない。
ファイル名が slug である。

### Issue

識別子は `SEN-n` である。
タイトル、本文、状態、優先度、Label、任意の Project、任意の Cycle、任意の期限、並び順、任意の親 Issue を持つ。
担当者は持たない。
親は高々1つである。深さの上限は設けない。循環は拒否する。

### メモ

Issue に紐づく本文と作成時刻だけを持つ。
編集と削除は提供しない。

### Page

タイトル、slug、本文、任意の親、任意の Project、状態、文書日付、tags を持つ。
ADR や自分用の文書に使う。

## UI

起動直後は `/issues` を出す。
左ナビ、中央のリストまたはボード、右の詳細である。

経路は `/issues`、`/board`、`/issues/{identifier}`、`/projects`、`/projects/{slug}`、`/cycles`、`/cycles/{number}`、`/views/{slug}`、`/pages`、`/pages/{slug}` である。
保存した View は左ナビに並ぶ。

キーボードは `Mod+K`、`c`、`p`、`j` / `k`、`Enter`、`Esc`、`s`、`1` から `4` である。
パレットから Issue 作成、画面移動、状態変更、Cycle への割り当てができる。

サブ Issue は親の詳細に一覧し、リストでは子を一段下げて示す。

## 同期

成果物の参照は `ghcr.io/<user>/sen` である。
自分のマシン間のバックアップであり、共有ディレクトリではない。
`sen pull` はローカルが dirty なら中止する。
起動時自動同期と `--force` pull は持たない。

## 代替案

Cycle をチームのスプリントと同一視して削除する案は採らない。
個人の時間枠として残す。

サブ Issue をブロック関係まで広げる案は採らない。
親子だけにする。
依存の表現は見積もりや担当者と結びつきやすいためである。

カスタム View をチーム共有の画面として入れる案は採らない。
自分の絞り込みの保存だけにする。

## 懸念点

親 Issue とカスタム View は実装済みである。

## 未決定事項

親子はリストで一段下げて示す。深さの上限は設けない。
View の絞り込みに全文検索 `q` は持たない。状態、Project、Cycle、Label、優先度だけである。
