// CLIエントリーポイント（将来実装予定）
//
// このCLIツールはHTTPサーバーを介さずにServiceを直接呼び出す。
// Handlerを通じてできる操作はすべてCLIからも実行できる。
//
// 使用例（将来）:
//
//	go run ./cmd/cli issue create --project=PROJ --title="バグ修正" --reporter=<uuid>
//	go run ./cmd/cli issue list --project=PROJ
//	go run ./cmd/cli project create --key=PROJ --name="マイプロジェクト" --owner=<uuid>
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "CLIツールは未実装です。Serviceレイヤー実装後に追加予定。")
	os.Exit(1)
}
