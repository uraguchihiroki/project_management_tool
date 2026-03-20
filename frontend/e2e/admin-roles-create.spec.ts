import { test, expect } from '@playwright/test'
import { setupAuth, TEST_ORG_ID } from './helpers'

test.describe('組織管理者: 役職作成', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page, 'e2e-roles@example.com')
  })

  test('フォームで name/level 入力 → 作成 → 一覧に表示', async ({ page }) => {
    const unique = Date.now()
    const name = `E2E役職${unique}`

    await page.goto('/admin/roles')
    await expect(page.getByRole('heading', { name: /役職/ })).toBeVisible({ timeout: 5000 })

    await page.getByRole('button', { name: '役職を追加' }).click()
    await page.getByPlaceholder('例: 部長').fill(name)
    await page.getByPlaceholder('1').fill('3')
    await page.getByRole('button', { name: '追加' }).click()

    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 })
  })
})
