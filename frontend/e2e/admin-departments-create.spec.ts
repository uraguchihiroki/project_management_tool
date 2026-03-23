import { test, expect } from '@playwright/test'
import { setupAuth } from './helpers'

test.describe('組織管理者: グループ作成', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page, 'e2e-depts@example.com')
  })

  test('フォームで name 入力 → 作成 → 一覧に表示', async ({ page }) => {
    const unique = Date.now()
    const name = `E2Eグループ${unique}`

    await page.goto('/admin/departments')
    await expect(page.getByRole('heading', { name: /グループ/ })).toBeVisible({ timeout: 5000 })

    await page.getByRole('button', { name: /グループを追加|新規/ }).first().click()
    await page.getByPlaceholder('例: 開発部、予算委員会（200文字以内）').fill(name)
    await page.getByRole('button', { name: /追加|作成/ }).click()

    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 })
  })
})
