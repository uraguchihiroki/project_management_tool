import { test, expect } from '@playwright/test'
import { ensureLoginableUser, E2E_LOGIN_EMAIL } from './helpers'

/**
 * 管理者チェック付きログイン後、/admin/projects（管理画面）まで到達できること。
 * （sessionStorage の is_admin はログイン時のチェックボックスで付与）
 */
test.describe('ログイン→管理画面', () => {
  test.beforeEach(async ({ page }) => {
    // addInitScript で毎回 clear すると、後続の goto('/admin/...') のたびに session が消えてログイン画面に戻る
    await page.goto('/login', { waitUntil: 'load' })
    await page.evaluate(() => {
      try {
        sessionStorage.clear()
        localStorage.clear()
      } catch {
        /* ignore */
      }
    })
    await page.reload({ waitUntil: 'load' })
    await expect(page.getByRole('heading', { name: 'ログイン' })).toBeVisible()
  })

  test('管理者としてログインし /admin/projects を表示できる', async ({ request, page }) => {
    await ensureLoginableUser(request)

    await page.getByPlaceholder('例: taro@example.com').fill(E2E_LOGIN_EMAIL)
    await page.getByRole('checkbox', { name: /管理者としてログイン/ }).check()

    const loginPost = page.waitForResponse(
      (r) =>
        r.url().includes('admin/login') &&
        r.request().method() === 'POST' &&
        r.ok(),
      { timeout: 30_000 },
    )
    await page.getByRole('button', { name: 'ログイン' }).click()
    await loginPost

    await page.waitForResponse(
      (r) =>
        r.url().includes('/users/') &&
        r.url().includes('/organizations') &&
        r.request().method() === 'GET' &&
        r.ok(),
      { timeout: 30_000 },
    )

    await expect(page).toHaveURL(/\/projects|\/select-org/, { timeout: 25_000 })

    // レイアウトは currentUser 復元まで「読み込み中…」のため、ヘッダー「管理画面」よりサイドバーを待つ
    await page.goto('/admin/projects', { waitUntil: 'domcontentloaded' })
    await expect(page.getByRole('link', { name: 'プロジェクト管理' })).toBeVisible({ timeout: 25_000 })
    await expect(page.getByText('管理画面', { exact: false })).toBeVisible({ timeout: 10_000 })

    await page.goto('/admin/workflows', { waitUntil: 'domcontentloaded' })
    await expect(page.getByRole('heading', { name: /ワークフロー/ })).toBeVisible({ timeout: 25_000 })
  })
})
