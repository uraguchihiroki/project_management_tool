import { test, expect } from '@playwright/test'
import { setupAuth, TEST_ORG_ID } from './helpers'

test.describe('組織管理者: ユーザー作成', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page, 'e2e-users@example.com')
  })

  test('フォームで name/email 入力 → 作成 → 一覧に表示', async ({ page }) => {
    const unique = Date.now()
    const name = `E2Eユーザー${unique}`
    const email = `e2e-user-${unique}@example.com`

    await page.goto('/admin/users')
    await expect(page.getByRole('heading', { name: /ユーザー/ })).toBeVisible({ timeout: 5000 })

    await page.getByRole('button', { name: 'ユーザー登録' }).click()
    await page.getByPlaceholder('山田 太郎').fill(name)
    await page.getByPlaceholder('taro@example.com').fill(email)
    await page.locator('form').getByRole('button', { name: /登録|作成|追加/ }).click()

    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 })
    await expect(page.getByText(email)).toBeVisible()
  })
})
