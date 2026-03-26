# プロジェクト編集

## パス

`/admin/projects/[id]`

## 概要

プロジェクトの詳細編集。名前、説明、開始日、終了日を変更。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

- 遷移元: /admin/projects
- 遷移先: 戻る → /admin/projects

## UI 要素

- 戻るリンク「プロジェクト管理に戻る」
- 見出し「プロジェクト編集」
- フォーム: プロジェクトキー（読取専用）、プロジェクト名、説明、開始日、終了日
- 保存ボタン

## 備考

- `GET /projects/:id` でプロジェクト取得
- `PUT /projects/:id` でプロジェクト更新
- プロジェクトキーは変更不可
