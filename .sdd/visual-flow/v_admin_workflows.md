# ワークフロー一覧

## パス

`/admin/workflows`

## 概要

ワークフローの一覧・作成・編集・削除。ワークフローは承認プロセスを定義する。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

- 遷移先: ワークフロークリック → /admin/workflows/[id]

## UI 要素

- 見出し「ワークフロー」
- ワークフロー追加ボタン
- ワークフロー追加/編集フォーム（名前、説明）
- ワークフロー一覧（ドラッグで並び替え、クリックで詳細へ）

## 備考

- `GET /workflows` でワークフロー一覧（ユーザーステップが1つ以上あるもののみ）
- `POST /workflows`、`PUT /workflows/:id`、`DELETE /workflows/:id`
- `PUT /workflows/reorder` で並び替え
