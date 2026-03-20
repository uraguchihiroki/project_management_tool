/** @type {import('next').NextConfig} */
const nextConfig = {
  // プロジェクトルートを明示（複数 lockfile 警告対策）
  turbopack: { root: __dirname },
  // 開発時の左下「Compiling / Rendering」表示を消す（ログイン後の遷移が止まって見える対策）
  // エラー時のオーバーレイは引き続き表示される
  devIndicators: false,
}

module.exports = nextConfig
