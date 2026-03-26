import { test, expect } from '@playwright/test'
import { setupAuth } from './helpers'

test.describe('組織管理者: プロジェクト作成', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page, 'e2e-projects@example.com')
  })

  test('フォームで key/name 入力 → 作成 → 一覧に表示', async ({ page }) => {
    const unique = Date.now().toString(36).slice(-4).toUpperCase()
    const key = `E2E${unique}`
    const name = `E2Eプロジェクト${unique}`

    await page.goto('/admin/projects')
    await expect(page.getByRole('heading', { name: /プロジェクト/ })).toBeVisible({ timeout: 5000 })

    await page.getByRole('button', { name: /新規プロジェクト/ }).first().click()
    await page.getByPlaceholder('例: PROJ').fill(key)
    await page.getByPlaceholder('プロジェクト名を入力').fill(name)
    await page.locator('form').getByRole('button', { name: /作成/ }).click()

    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 })
    await expect(page.getByText(key)).toBeVisible()
  })
})
