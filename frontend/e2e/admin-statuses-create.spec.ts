import { test, expect } from '@playwright/test'
import { setupAuth } from './helpers'

test.describe('組織管理者: ステータス作成', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page, 'e2e-statuses@example.com')
  })

  test('フォームで name/color 入力 → 作成 → 一覧に表示', async ({ page }) => {
    const unique = Date.now()
    const name = `E2Eステータス${unique}`

    await page.goto('/admin/statuses')
    await expect(page.getByRole('heading', { name: /ステータス/ })).toBeVisible({ timeout: 5000 })

    await page.getByRole('button', { name: 'ステータスを追加' }).click()
    await page.getByPlaceholder('例: 未着手').fill(name)
    await page.getByPlaceholder('#6B7280').fill('#3B82F6')
    await page.getByRole('button', { name: '追加' }).click()

    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 })
  })
})
