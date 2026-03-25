# ステータス管理（レガシー導線）

## パス

`/admin/statuses`（現在は `/admin/workflows` へリダイレクト）

## 概要

この導線は後方互換のために残し、実運用では **ワークフロー詳細画面（`/admin/workflows/[id]`）で Issue ステータスを編集**する。

## 遷移元・遷移先

[transition-flow.md](transition-flow.md) を参照。

## 現在の挙動

- `/admin/statuses` へアクセスすると `/admin/workflows` へリダイレクト
- Issue ステータスの追加・編集は `/admin/workflows/[id]` の **同一ダイアログ** で実施（新規/編集を統一）
- ワークフロー詳細では **ステータス一覧テーブル内**で **開始（常に1件・ラジオで付け替え）・終了（複数）** をラジオ／チェックで設定し、遷移図に START / GOAL マークを表示する（新規3列ブートストラップの既定開始は **表示順が最小の列**）

## 備考

- API 自体は `GET /organizations/:orgId/statuses`、`POST /organizations/:orgId/statuses`、`PUT /statuses/:id`、`DELETE /statuses/:id`
- UI からの更新主導線はワークフロー詳細
