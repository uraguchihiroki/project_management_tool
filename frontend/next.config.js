/** @type {import('next').NextConfig} */
const nextConfig = {
  // プロジェクトルートを明示（複数 lockfile 警告対策）
  turbopack: { root: __dirname },
}

module.exports = nextConfig
