# スーパー管理者 / 管理者ログイン

## パス

`/super-admin/login`

## 概要

スーパー管理者専用のログイン画面。メールアドレスのみでログイン。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

- 遷移先: ログイン成功時 → /super-admin

## UI 要素

- ロゴ（Shield アイコン + 「スーパー管理者」）
- 見出し「管理者ログイン」
- メールアドレス入力
- ログインボタン
- 通常ユーザーログインへのリンク（/login）

## 備考

- `POST /super-admin/login` で認証
- `super_admins` テーブルに存在するメールのみログイン可能
- 初期アカウント: seed.sql 実行後、`superadmin@frs.example.com`
