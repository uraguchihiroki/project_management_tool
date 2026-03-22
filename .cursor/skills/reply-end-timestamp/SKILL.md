---
name: reply-end-timestamp
description: >-
  エージェントの返信末尾に応答時点の日付と時刻の両方を1行付ける（形式: 応答日時 YYYY-MM-DD HH:MM JST）。
  全プロジェクト適用は Cursor User Rules。本リポでは .cursor/rules/reply-end-timestamp.mdc も alwaysApply。
---

# 返信末尾に日時を付ける

## 優先順位（どこに書くか）

1. **全 Cursor プロジェクト** … **Cursor Settings → User Rules** にルールを貼る（手順は [docs/cursor-user-rules-reply-timestamp.md](../../../docs/cursor-user-rules-reply-timestamp.md)）。
2. **本リポジトリ** … [`.cursor/rules/reply-end-timestamp.mdc`](../../rules/reply-end-timestamp.mdc) が `alwaysApply: true` で補強する。

## ルール

1. **ユーザーへの返信**（説明・まとめ・レビュー・手順案内など、実質的な文面があるもの）の **最後の行** に、応答時点の **日付と時刻の両方** を付ける。
2. **形式**（推奨）: `（応答日時: YYYY-MM-DD HH:MM JST）`
3. **日付の出所**: メッセージに含まれる user_info の `Today's date` があればそれを使う。
4. **時刻**: user_info に時刻が無い場合は、正確に付けるために **1 回** ターミナルで  
   `TZ=Asia/Tokyo date '+%Y-%m-%d %H:%M %Z'`  
   を実行してよい（WSL の bash 想定。Windows 版 Cursor + Remote WSL では統合ターミナルが bash になりやすい）。
5. **例外**: ツール呼び出しのみ・1 語の返答など、末尾に時刻がノイズになるほど短い返信では省略してよい。

## ソースコード内のコメント（任意）

長めのブロックコメントや TODO を書くとき、**末尾に記述日** を付けてもよい（例: `// 2026-03-22 方針メモ`）。必須ではない。

## 例

**悪い（日時なし）**

> 修正しました。`foo.go` を更新しています。

**良い**

> 修正しました。`foo.go` を更新しています。
>
> （応答日時: 2026-03-22 14:05 JST）
