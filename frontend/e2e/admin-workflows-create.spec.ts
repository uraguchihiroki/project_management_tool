import { test, expect } from '@playwright/test'
import { setupAuth } from './helpers'

test.describe('組織管理者: ワークフロー作成', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page, 'e2e-workflows@example.com')
  })

  test('フォームで name/description 入力 → 作成 → 一覧に表示', async ({ page }) => {
    const unique = Date.now()
    const name = `E2Eワークフロー${unique}`

    await page.goto('/admin/workflows')
    await expect(page.getByRole('heading', { name: /ワークフロー/ })).toBeVisible({ timeout: 5000 })

    // addInitScript で入れた JWT / currentOrg が React に反映されるまで待つ
    await page.waitForFunction(
      () =>
        sessionStorage.getItem('authToken') != null &&
        sessionStorage.getItem('currentOrg') != null &&
        sessionStorage.getItem('currentOrg') !== '',
      { timeout: 15_000 },
    )
    await page.getByRole('button', { name: /ワークフローを追加/ }).click()
    await page.getByPlaceholder('例: 通常Issue').fill(name)
    await page.getByPlaceholder('例: 一般的な業務申請に使用').fill('E2Eテスト用')
    // h2「新しいワークフロー」は form の外にあるため、送信ボタンは role で直接指定する
    await page.getByRole('button', { name: '追加' }).click()

    await expect(page.getByText(name)).toBeVisible({ timeout: 10_000 })
  })
})
