# Cursor：返信末尾に日付・時刻を付ける（User Rules 設定手順）

**Windows 版 Cursor** で **WSL 上のプロジェクト**（例: `~/work/AI/project_management_tool`）を開いていても、**全プロジェクト共通のルール**は **Windows 側の Cursor 設定**に書きます。

- **User Rules** … Cursor 専用の **Cursor Settings** 画面にあります（後述）。**Git には含まれません**。  
  ※昔の説明やコミュニティでは「Rules for AI」と呼ばれることがありましたが、**現行 UI では「User Rules」や「Rules」内のグローバル欄**として表示されることが多いです。
- **プロジェクトルール** … 本リポジトリでは [`.cursor/rules/reply-end-timestamp.mdc`](../.cursor/rules/reply-end-timestamp.mdc) もあります（このフォルダを開いたときの補強）。

---

## 手順（ユーザーが行うこと）

### 1. 開くべき画面は「Cursor Settings」（VS Code の「設定」ではない）

**よくある勘違い**: **`Ctrl + ,`（カンマ）** や **`File` → `Preferences` → `Settings`** で開くのは、主に **VS Code 互換のエディタ設定**です。ここには **User Rules の欄はありません**。

**開くべきなのは「Cursor Settings」**（Cursor 独自の設定 UI）です。次を順に試してください。

1. **`Ctrl + Shift + P`**（コマンドパレット）を開く。
2. 次のいずれかを入力して実行する（表示される方）:
   - `Cursor Settings`
   - `Preferences: Open Cursor Settings`
   - `open cursor settings`（英語 UI の場合）
3. 別の方法: メニュー **`File` → `Preferences` → `Cursor Settings`**（項目があればここ）。
4. 別の方法: ショートカット **`Ctrl + Shift + J`**（環境によっては Cursor Settings が開く）。

開けた画面の **左側ナビ**で次を探します。

- **`Rules`** という項目、または
- **`Rules, Commands`**（プロジェクトルールと並んで、**グローバル／User** 用のテキストエリアがあることが多い）

その中の **User Rules**（または「For all projects」「Always applied」に近い、**リポジトリに依存しない**説明文付きの入力欄）に本文を貼ります。

**まだ見つからない場合**

- Cursor のバージョンでラベルが違います。コマンドパレットで **`rules`** とだけ打ち、**設定を開く系のコマンド**がないか一覧から探してください。
- それでも無い場合は **このリポジトリだけ**でも [`.cursor/rules/reply-end-timestamp.mdc`](../.cursor/rules/reply-end-timestamp.mdc)（`alwaysApply: true`）が効きます。**全プロジェクト**まで揃えるには、Cursor 公式ドキュメントの [Rules](https://cursor.com/docs/context/rules) の **User Rules** の項を開き、自分の版に合わせてください。

### 2. 下記のブロックをコピーして貼り付け

`User Rules` の欄に、**次の全文をそのまま**貼り付けて保存する。

```text
最優先: このターンでユーザーに見える返信テキストが1文字でもあるなら、必ず最後の行に応答時点の日付と時刻の両方を付ける。Plan の本文・説明・手順・箇条書きも含む。

形式: （応答日時: YYYY-MM-DD HH:MM JST）

日付はメッセージの user_info に Today's date があればそれを使う。時刻がコンテキストに無い場合は、正確に付けるためにターミナルで TZ=Asia/Tokyo date '+%Y-%m-%d %H:%M %Z' を1回実行してよい（WSL の bash 想定）。

省略してよいのは、ユーザー向けの自然文・見出し・箇条書きが一切無いターンのみ（ツールだけで本文ゼロ）。「はい」「OK」だけは省略可。
```

### 3. 動作確認

1. **別のフォルダ**を Cursor で開く（空のフォルダでも可）か、新規チャットを開始する。
2. AI に短い質問をして、返信の**最後の行**に  
   `（応答日時: YYYY-MM-DD HH:MM JST）`  
   の形で**日付と時刻の両方**が付くか確認する。

---

## 参考：設定ファイルの場所（閲覧用・手編集非推奨）

User Rules の実体は、多くの場合 **Windows ユーザーデータ**配下にあります。

| 環境 | 典型的なルート（フルパス例） |
|------|------------------------------|
| Windows 版 Cursor | `C:\Users\<Windowsのユーザー名>\AppData\Roaming\Cursor\` |

その下に `User\settings.json` や `User\globalStorage\` などがあることがありますが、**中身は Cursor のバージョンで変わる**ため、**必ず設定 UI から編集**してください。

**注意**: プロジェクトが WSL 上（例: `/home/uraguchi/work/AI/project_management_tool`）でも、**User Rules は Linux の `~/.config/Cursor` ではなく、基本は上記 Windows 側**です。

---

## エージェント向け（詳細・例）

[`.cursor/skills/reply-end-timestamp/SKILL.md`](../.cursor/skills/reply-end-timestamp/SKILL.md) に例や補足があります。
