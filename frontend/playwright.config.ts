import { defineConfig, devices } from '@playwright/test'

const wsEndpoint = process.env.PLAYWRIGHT_WS_ENDPOINT

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  // 既定 test-results が権限で書けない環境向け（.gitignore 済み）
  outputDir: process.env.PLAYWRIGHT_OUTPUT_DIR || 'test-results-local',
  reporter: 'html',
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:3000',
    trace: 'on-first-retry',
    ...(wsEndpoint ? { connectOptions: { wsEndpoint } } : {}),
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  // 実行前に backend と frontend を起動してください
  // backend: go build -o server.exe ./cmd/server && .\server.exe
  // frontend: npm run dev
  //
  // Windows 側で `npx playwright run-server --port 9222` 済みなら、
  // WSL 側で PLAYWRIGHT_WS_ENDPOINT=ws://<windows-host-ip>:9222/ を渡すと
  // ローカル起動の代わりに Playwright Server へ接続して実行できます。
})
