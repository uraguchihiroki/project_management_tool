import { test, expect } from '@playwright/test'

const API = process.env.PLAYWRIGHT_API_URL || 'http://localhost:8080/api/v1'

test.describe('スーパー管理者: 組織作成', () => {
  test.beforeEach(async ({ page }) => {
    // スーパーアドミンでログイン
    const loginRes = await fetch(`${API}/super-admin/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: 'superadmin@frs.example.com' }),
    })
    if (loginRes.status !== 200) {
      test.skip(true, 'スーパーアドミンが seed に存在しないためスキップ')
    }
    const loginJson = await loginRes.json()
    const sa = loginJson.data
    if (!sa) test.skip(true, 'スーパーアドミンログイン失敗')

    await page.addInitScript(
      ({ saData }: { saData: { id: string; name: string; email: string } }) => {
        sessionStorage.setItem('currentSuperAdmin', JSON.stringify(saData))
      },
      { saData: sa }
    )
  })

  test('フォームで組織名・管理者情報入力 → 作成 → 一覧に表示', async ({ page }) => {
    const unique = Date.now()
    const name = `E2E組織${unique}`

    await page.goto('/super-admin')
    await expect(page.getByRole('heading', { name: /会社・組織|組織/ })).toBeVisible({ timeout: 5000 })

    await page.getByRole('button', { name: /新規会社作成/ }).click()
    await page.getByPlaceholder(/組織名|会社名/).fill(name)
    await page.locator('form').getByRole('button', { name: /作成/ }).click()

    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 })
  })
})
