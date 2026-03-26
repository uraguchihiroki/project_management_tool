# プロジェクト一覧

## パス

`/projects`

## 概要

組織に紐づくプロジェクトの一覧を表示。プロジェクトをクリックで詳細へ遷移。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

- 遷移元: /login、/select-org、/admin
- 遷移先: プロジェクトクリック → /projects/[id]

## UI 要素

- 見出し「プロジェクト一覧」
- プロジェクトカード一覧（キー、名前、説明等）
- ヘッダー: 組織名、管理画面リンク、ユーザー情報

## 備考

- `GET /projects?org_id=xxx` でプロジェクト一覧取得
- 組織選択済みが前提
