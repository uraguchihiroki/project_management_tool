---
name: reply-end-timestamp
description: >-
  エージェントの返信末尾に応答時点の日付と時刻の両方を1行付ける（形式: 応答日時 YYYY-MM-DD HH:MM JST）。
  全プロジェクト適用は Cursor User Rules。本リポでは .cursor/rules/reply-end-timestamp.mdc が alwaysApply で最優先。
---

# 返信末尾に日時を付ける

## 優先順位（どこに書くか）

1. **全 Cursor プロジェクト** … **Cursor Settings → User Rules** にルールを貼る（手順は [docs/cursor-user-rules-reply-timestamp.md](../../../docs/cursor-user-rules-reply-timestamp.md)）。
2. **本リポジトリ** … [`.cursor/rules/reply-end-timestamp.mdc`](../../rules/reply-end-timestamp.mdc) が `alwaysApply: true` で **最優先**。

## ルール（mdc と同一方針）

1. **ユーザーに見える返信テキストが1文字でもあるターン**は、**最後の行**に応答時点の **日付と時刻の両方** を付ける。Plan・説明・手順・箇条書きも含む。
2. **形式**: `（応答日時: YYYY-MM-DD HH:MM JST）`
3. **日付**: user_info の `Today's date` があればそれを使う。
4. **時刻**: 無い場合は **1 回** ターミナルで  
   `TZ=Asia/Tokyo date '+%Y-%m-%d %H:%M %Z'`  
   を実行してよい（WSL bash 想定。Windows 版 Cursor + Remote WSL では統合ターミナルが bash になりやすい）。
5. **省略可**: **ユーザー向けの自然文・見出し・箇条書きが一切無い**ターンのみ。「はい」「OK」だけは省略可（付けてもよい）。

## ソースコード内のコメント（任意）

長めのブロックコメントや TODO を書くとき、**末尾に記述日** を付けてもよい（例: `// 2026-03-22 方針メモ`）。必須ではない。

## 例

**悪い（日時なし）**

> 修正しました。`foo.go` を更新しています。

**良い**

> 修正しました。`foo.go` を更新しています。
>
> （応答日時: 2026-03-22 14:05 JST）
