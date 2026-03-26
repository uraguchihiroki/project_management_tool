import { test, expect } from '@playwright/test'
import { ensureLoginableUser, E2E_LOGIN_EMAIL } from './helpers'

/**
 * ログイン画面のスモーク（API＋UI 結線）。
 *
 * 前提: バックエンドが `PLAYWRIGHT_API_URL`（既定: http://localhost:8080/api/v1）で起動していること。
 * フロントは `PLAYWRIGHT_BASE_URL`（既定: http://localhost:3000）。
 *
 * 実行例:
 *   cd frontend && npm run test:e2e -- e2e/login.spec.ts
 */
test.describe('ログイン', () => {
  test.beforeEach(async ({ page }) => {
    // reload は React の入力と競合し「email is required」になりやすいので使わない
    await page.addInitScript(() => {
      try {
        sessionStorage.clear()
        localStorage.clear()
      } catch {
        /* ignore */
      }
    })
    await page.goto('/login', { waitUntil: 'load' })
    await expect(page.getByRole('heading', { name: 'ログイン' })).toBeVisible()
  })

  test('登録済みメールでログインし、プロジェクト一覧または組織選択へ遷移する', async ({ request, page }) => {
    await ensureLoginableUser(request)

    const emailInput = page.getByPlaceholder('例: taro@example.com')
    await emailInput.fill(E2E_LOGIN_EMAIL)
    await emailInput.blur()
    await expect(emailInput).toHaveValue(E2E_LOGIN_EMAIL)

    const loginPost = page.waitForResponse(
      async (r) => {
        if (!r.url().includes('admin/login') || r.request().method() !== 'POST') return false
        try {
          const body = r.request().postDataJSON() as { email?: string } | null
          return body?.email === E2E_LOGIN_EMAIL
        } catch {
          return false
        }
      },
      { timeout: 30_000 },
    )
    await page.getByRole('button', { name: 'ログイン' }).click()
    const loginRes = await loginPost
    if (!loginRes.ok()) {
      throw new Error(`admin/login failed: ${loginRes.status()} ${await loginRes.text()}`)
    }

    // JWT 後の所属組織取得が終わるまで待つ（遷移前に失敗すると /login のまま）
    await page.waitForResponse(
      (r) =>
        r.url().includes('/users/') &&
        r.url().includes('/organizations') &&
        r.request().method() === 'GET' &&
        r.ok(),
      { timeout: 30_000 },
    )

    await expect(page).toHaveURL(/\/projects|\/select-org/, { timeout: 25_000 })
    await expect(page.getByTestId('login-error')).toHaveCount(0)
  })

  test('未登録メールではエラー表示され、ログイン完了しない', async ({ page }) => {
    // ensureLoginableUser は不要（API が 401 を返せればよい）
    const unknown = `e2e-no-user-${Date.now()}@example.com`

    const unknownInput = page.getByPlaceholder('例: taro@example.com')
    await unknownInput.fill(unknown)
    await unknownInput.blur()
    await expect(unknownInput).toHaveValue(unknown)
    await page.getByRole('button', { name: 'ログイン' }).click()

    await expect(page.getByRole('heading', { name: 'ログイン' })).toBeVisible({ timeout: 5000 })
    const err = page.getByTestId('login-error')
    await expect(err).toBeVisible({ timeout: 20_000 })
    await expect(err).toContainText(
      /メールアドレスが見つかりません|ログインに失敗|APIに接続|エラーが発生しました|所属組織|email is required/i,
    )
    await expect(page).toHaveURL(/\/login/)
  })
})
