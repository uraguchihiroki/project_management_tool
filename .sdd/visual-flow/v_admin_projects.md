# プロジェクト管理

## パス

`/admin/projects`

## 概要

組織のプロジェクトの作成・一覧・並び替え。プロジェクト詳細（編集）へ遷移。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

- 遷移先: プロジェクトクリック → /admin/projects/[id]

## UI 要素

- 見出し「プロジェクト管理」、説明
- プロジェクト追加フォーム（キー、名前、説明、開始日、終了日）
- プロジェクト一覧（ドラッグで並び替え、クリックで編集へ）

## 備考

- `GET /projects?org_id=xxx` でプロジェクト一覧
- `POST /projects` でプロジェクト作成
- `PUT /projects/reorder` で並び替え
