# プロジェクト詳細

## パス

`/projects/[id]`

## 概要

プロジェクト内の Issue 一覧をカンバン形式で表示。Issue の作成・ステータス変更が可能。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

- 遷移元: /projects
- 遷移先: Issue クリック → /projects/[id]/issues/[number]、戻る → /projects

## UI 要素

- ヘッダー: プロジェクト名、戻るリンク
- カンバン: ステータスごとの列、Issue カード
- Issue 作成フォーム（テンプレート選択、タイトル、説明、ステータス、優先度、担当者）
- テンプレート選択ドロップダウン

## 備考

- `GET /projects/:id` でプロジェクト・ステータス取得
- `GET /projects/:projectId/issues` で Issue 一覧
- `POST /projects/:projectId/issues` で Issue 作成
- テンプレート選択時は本文・優先度等を自動適用
