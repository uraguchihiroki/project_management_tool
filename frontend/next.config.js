/** @type {import('next').NextConfig} */
// Playwright を Windows ホストで動かし WSL で Next を動かす場合、ブラウザ内の fetch は
// localhost:8080 が「Windows 側」になり API に届かない。同一オリジン /api/v1 を Next が
// WSL 上のバックエンドへプロキシする。
const backendProxyTarget = process.env.BACKEND_PROXY_TARGET || 'http://127.0.0.1:8080'

const nextConfig = {
  // ブラウザが /favicon.ico を直接取りに行くため、SVG へ誘導（404 抑止）
  async redirects() {
    return [
      {
        source: '/favicon.ico',
        destination: '/favicon.svg',
        permanent: false,
      },
    ]
  },
  async rewrites() {
    return [
      {
        source: '/api/v1/:path*',
        destination: `${backendProxyTarget.replace(/\/+$/, '')}/api/v1/:path*`,
      },
    ]
  },
  // プロジェクトルートを明示（複数 lockfile 警告対策）
  turbopack: { root: __dirname },
  // 開発時の左下「Compiling / Rendering」表示を消す（ログイン後の遷移が止まって見える対策）
  // エラー時のオーバーレイは引き続き表示される
  devIndicators: false,
}

module.exports = nextConfig
