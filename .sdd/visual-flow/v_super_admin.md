# スーパー管理者パネル

## パス

`/super-admin`

## 概要

スーパー管理者専用の管理画面。組織の作成・一覧のみ可能。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

- 遷移元: /super-admin/login

## UI 要素

- ヘッダー: 「スーパー管理者パネル」、ユーザー名、ログアウトボタン
- 組織一覧
- 組織作成フォーム（組織名、管理者メール、管理者名）

## 備考

- `GET /super-admin/organizations` で組織一覧取得
- `POST /super-admin/organizations` で組織作成
- 組織作成時に指定メールのユーザーが組織管理者として追加される
