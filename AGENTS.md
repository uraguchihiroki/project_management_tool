# Issue Management Tool — AI Assistant Context

**Issue 管理**を主目的とした、Jira / Redmine ライクなチケット・カンバン型のツール。マルチテナント対応（組織ごとにデータ分離）。ステータス遷移の権限は [.sdd/transition-permissions.md](.sdd/transition-permissions.md) を参照。

---

## 最重要事項（人・AI ともに必ず認識すること）

### 1. 開発はローカル、運用はクラウド

**このプロジェクトは開発をローカル PC で行い、運用は GCP などのクラウドで行う。**

- AI はこの前提を常に考えて提案すること
- プランニング・設計・実装のすべてにおいて最上位で考慮する
- 詳細は [.sdd/principles.md](.sdd/principles.md) を参照

### 2. WSL にプロジェクトを置いた理由

- **経緯**: もともと Windows 側で開発していた
- **問題**: 度重なる Docker Desktop の不具合と、PowerShell と AI の相性の悪さで開発効率が著しく低下した
- **対応**: WSL 側に環境を構築し、プロジェクトを WSL 上に移した

### 3. Windows 版 Cursor を使う場合の必須条件

**WSL にアクセスする機能拡張（Remote - WSL）を使って、必ず WSL 上のプロジェクトに接続すること。**

- **経緯**: 当初、Ubuntu シェルから `/mnt/c/.../cursor.exe` のように Cursor を起動していた
- **問題**: その方法だと Cursor を落とすたびに記憶喪失（チャット履歴・コンテキストの消失）が発生した
- **解決**: WSL 接続の機能拡張を追加したうえで、WSL 上のプロジェクトを開くことで記憶喪失から抜け出せた
- **結論**: Windows 版 Cursor を使う場合は、**必ず WSL に接続してからプロジェクトを開く**こと

### 4. コミットメッセージ言語ポリシー（最上位ルール）

**AI が作成するコミットメッセージは、必ず日本語で記述すること。**

- 英語コミットメッセージは禁止（ユーザーが明示的に英語を要求した場合を除く）
- Conventional Commits のプレフィックス（`feat:`, `fix:` など）は使用可だが、本文は日本語にする
- このルールはコミット文面に関する他の慣習より優先する

### 5. エージェント返信末尾の日時（日付＋時刻）

**実質的な文面があるユーザー向け返信の最後の行に、日付と時刻の両方を付ける。**

- **各ターンの「ユーザーに見える最終メッセージ」に必ず付ける**（Plan 本文・説明のみの返信も含む。省略できるのは本文ゼロのターンのみ）。詳細は [`.cursor/rules/reply-end-timestamp.mdc`](.cursor/rules/reply-end-timestamp.mdc)。
- **全 Cursor プロジェクト**で同じにしたい場合の正本は **Cursor User Rules**（Windows 版 Cursor では Windows 側の設定）。**AI がユーザーの Cursor 設定を代わりに書き換えることはできない**ため、人が **1 回** 貼り付ける。手順とコピペ用テキスト → [docs/cursor-user-rules-reply-timestamp.md](docs/cursor-user-rules-reply-timestamp.md)
- **本リポジトリを開いているとき**は [`.cursor/rules/reply-end-timestamp.mdc`](.cursor/rules/reply-end-timestamp.mdc)（`alwaysApply: true`）も読み込まれる。
- 形式・`date` コマンド・例 → [`.cursor/skills/reply-end-timestamp/SKILL.md`](.cursor/skills/reply-end-timestamp/SKILL.md)

### 6. ユーザーからの不具合・挙動の問い合わせ（AI の調査順）

**問い合わせの多くは、ユーザーがブラウザで操作して見えた結果に基づく。** だから **ログインに限らず**、画面・遷移・エラー表示などの報告では、**コードや DB をいきなり疑わず、まず自分もブラウザで事実を確認する。**

**必ず先に（可能な範囲で）やること**:

1. **ブラウザで再現** … 利用可能なら **Cursor の Browser MCP** で、ユーザーが言及した **URL・画面・操作**を追い、**自分の目で**表示やエラーを確認する（ユーザーと同じ土俵）。  
2. **サーバー疎通・再起動** … バックエンド（例: `GET /api/v1/health` が **200**）、フロント（例: `:3000` が応答）を確認。コード・`git pull`・設定変更のあと **再起動したか**も確認または聞く（Go は Django のように自動で載せ替わらない）。詳細は [.sdd/dev-guide.md](.sdd/dev-guide.md)。  
3. **ブラウザが使えないとき** … その旨を明示し、`curl`・E2E（[`.sdd/testing.md`](.sdd/testing.md)）などで **同じ失敗が再現するか** 代替する。

**これらのあと**で、Network・ログ、必要なら Handler / DB / プロキシを追う。常時ルールの詳細 → [`.cursor/rules/browser-first-investigation.mdc`](.cursor/rules/browser-first-investigation.mdc)。

**認証・管理画面まわりを実装・修正したあと人に渡す前**（AI が自律してできる範囲）: **疎通込みで** `bash scripts/verify-login-e2e.sh` を実行する（DB 起動・`go run`・`npm run dev`・Playwright まで一括）。`npx playwright run-server` は **ブラウザ用**であり API ではない。手順・Windows 側サーバーとの組み合わせは [`.sdd/testing.md`](.sdd/testing.md)「Playwright Server」。

**設計との矛盾**: ユーザーの説明や求める修正が、[`.sdd/tenant-invariants.md`](.sdd/tenant-invariants.md) や [`.sdd/api-spec.md`](.sdd/api-spec.md) と食い違うときは、**いきなりコードを書かず**不変条件を読み直し、必要ならユーザーに確認してから進める。

### 7. 実装と仕様の乖離を検知したとき

**実装を読んだ結果、仕様（主に `.sdd`）とコードの挙動が一致しないと分かったときは、ユーザーに確認する。** 実装を勝手に正とせず、仕様を更新するか実装を直すかをユーザーと決める。

**ユーザーが合意した乖離** だけを [.sdd/spec-implementation-divergences.md](.sdd/spec-implementation-divergences.md) に列挙する（未合意のズレは載せない）。§6 は依頼文面・不具合と設計書の矛盾、§7 は **コードと `.sdd` の差分の扱い** を想定する。

---

## ユーザースキルの参照

タスク実行前に、以下のディレクトリ内のスキルを必ず確認すること。

| コンテキスト | パス |
|-------------|------|
| **本リポジトリ（共有）** | `.cursor/skills/`（例: `reply-end-timestamp`）。常時ルールは `.cursor/rules/reply-end-timestamp.mdc`、**不具合報告は先にブラウザで確認**する順序は `.cursor/rules/browser-first-investigation.mdc` |
| Windows | `C:\Users\diversion\.cursor\skills\` |
| WSL | `/mnt/c/Users/diversion/.cursor/skills/` |

後者は同一の実体を指す。該当するスキルがあれば SKILL.md の手順に従う。

---

## 設計原則

**開発はローカル、運用はGCPなどのクラウド。** プランニング・設計において常に最上位で考慮すること。詳細は [.sdd/principles.md](.sdd/principles.md) を参照。

## 技術スタック

| 役割 | 技術 |
|------|------|
| フロントエンド | Next.js 14 (React / TypeScript / Tailwind CSS) |
| バックエンド | Go 1.22 + Echo v4 + GORM |
| データベース | PostgreSQL 16 |
| 認証 | メールアドレスのみ（JWT 未実装） |

## ディレクトリ構成

```
project_management_tool/
├── .sdd/              # 設計ドキュメント（必読）
├── backend/            # Go API サーバー
│   ├── cmd/server/     # エントリポイント
│   ├── internal/      # handler, service, repository, model
│   ├── test/          # ブラックボックステスト
│   └── seed.sql       # 初期データ投入スクリプト
├── frontend/           # Next.js
│   └── src/app/       # App Router ページ
├── scripts/            # 開発補助（起動・E2E 用ラッパー等）
└── docker-compose.yml  # PostgreSQL
```

## 設計ドキュメント

**必ず [.sdd/README.md](.sdd/README.md) を参照すること。** ナビゲーションと各ドキュメントの役割が記載されている。

- [architecture.md](.sdd/architecture.md) — システム構成・マルチテナント
- [tenant-invariants.md](.sdd/tenant-invariants.md) — テナント不変条件の正本（JWT・親子 API・禁止事項）
- [layer-responsibility.md](.sdd/layer-responsibility.md) — Handler / Service / Repository の責務
- [db-schema.md](.sdd/db-schema.md) — テーブル定義
- [api-spec.md](.sdd/api-spec.md) — REST API 仕様
- [key-flows.md](.sdd/key-flows.md) — 認証・組織・ステータス遷移の権限（主要フロー）
- [transition-permissions.md](.sdd/transition-permissions.md) — ステータス遷移の権限（候補比較・TBD）
- [dev-guide.md](.sdd/dev-guide.md) — 新機能追加の手順
- [spec-implementation-divergences.md](.sdd/spec-implementation-divergences.md) — 仕様と実装の乖離（ユーザー合意済みのみ）

## WSL 開発環境（エージェント向け）

本プロジェクトは **Windows で Cursor を起動し、WSL にリモート接続**して開発する構成を推奨する。

- **Cursor**: Windows で起動 → Remote - WSL で接続 → プロジェクトフォルダを開く
- **ターミナル**: WSL Ubuntu の bash を使用（PowerShell は使用しない）
- **AI のターミナル実行**: ユーザー方針として **権限はすべて許可（サンドボックスなし）**で実行する（詳細は [.cursor/rules/wsl-dev-environment.mdc](.cursor/rules/wsl-dev-environment.mdc)）
- **パス**: WSL 形式（例: `/home/uraguchi/work/AI/project_management_tool`）を使用する

エージェントがターミナルコマンドを実行する場合、WSL を明示的に呼び出す例:

```bash
wsl -d Ubuntu-22.04 -e bash -c "cd /home/uraguchi/work/AI/project_management_tool && 実行したいコマンド"
```

Cursor が WSL に接続済みの場合は、統合ターミナルは自動的に bash になるため、そのままコマンドを実行してよい。

## 開発時の注意

0. **不具合の相談を受けたら** … 実装を読む前に **最重要事項の §6** と [`.cursor/rules/browser-first-investigation.mdc`](.cursor/rules/browser-first-investigation.mdc)（**ブラウザで再現 → 疎通・再起動**）。仕様と実装の乖離に気づいたら **§7** と [.sdd/spec-implementation-divergences.md](.sdd/spec-implementation-divergences.md)。
1. **レイヤー責務を守る**: Handler は HTTP の入出力のみ。ビジネスロジックは Service、DB 操作は Repository。
2. **テストを実行**: `cd backend && go test ./test/... -v` でブラックボックステストを実行。E2E（Playwright）の前提・**Playwright Server（Windows 側ブラウザ + WSL から接続）**は [.sdd/testing.md](.sdd/testing.md) を参照。
3. **`backend/test` が失敗したとき**: **安易にテストだけ期待値を合わせない**。判断は [.sdd/testing.md](.sdd/testing.md) の「仕様との関係とテスト修正時の判断」に従う（実装修正か、仕様・ユーザー確認か）。
4. **API / DB 変更時**: `.sdd/api-spec.md` および `.sdd/db-schema.md` を更新する。
5. **アプリ起動**（WSL 環境）: `bash scripts/start.sh` で DB・バックエンド・フロントエンドを一括起動。または `docker compose up -d db` の後、`go run ./cmd/server` と `npm run dev` を別ターミナルで実行。
6. **再起動の目安**（Django の自動リロードと違う）: **バックエンドは `go run` がホットスワップしない**ので、`.go` 変更や `git pull` のあとは **必ず再起動**。フロントは `npm run dev` で多くの UI は HMR されるが、**`next.config.js` / 環境変数**を変えたら再起動。**挙動がおかしいときは両方止めてから `start.sh` し直す**のが早い。詳細は [.sdd/dev-guide.md](.sdd/dev-guide.md) の「いつバックエンド／フロントを再起動するか」。
