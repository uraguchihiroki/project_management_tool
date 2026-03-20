# Issue 詳細

## パス

`/projects/[id]/issues/[number]`

## 概要

Issue の詳細表示。ステータス変更、コメント投稿、承認操作が可能。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

- 遷移元: /projects/[id]（プロジェクト詳細のカンバン）
- 遷移先: 戻る → /projects/[id]

## UI 要素

- ヘッダー: Issue 番号・タイトル、戻るリンク
- ステータス表示・変更
- 優先度、担当者、起票者、期日
- 説明文
- コメント一覧・コメント投稿フォーム
- 承認セクション（承認オブジェクトごとの状態、承認/却下ボタン）

## 備考

- `GET /projects/:projectId/issues/:number` で Issue 取得
- `PUT /projects/:projectId/issues/:number` でステータス等更新
- `GET /issues/:issueId/comments`、`POST /issues/:issueId/comments` でコメント
- `GET /issues/:issueId/approvals`、`POST /approvals/:id/approve`、`POST /approvals/:id/reject` で承認
