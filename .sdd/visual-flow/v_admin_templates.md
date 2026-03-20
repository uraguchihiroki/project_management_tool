# Issueテンプレート管理

## パス

`/admin/templates`

## 概要

Issue 作成時に選択できるテンプレートを定義。ワークフローを紐づけ可能。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

## UI 要素

- 見出し「Issueテンプレート管理」、説明
- テンプレート追加ボタン
- テンプレート追加/編集フォーム（テンプレート名、説明、本文、デフォルト優先度、ワークフロー）
- テンプレート一覧（ドラッグで並び替え）

## 備考

- `GET /projects/:projectId/templates` でテンプレート一覧
- `POST /templates`、`PUT /templates/:id`、`DELETE /templates/:id`
- `PUT /projects/:projectId/templates/reorder` で並び替え
- テンプレート選択時、Issue にワークフローが自動適用される
