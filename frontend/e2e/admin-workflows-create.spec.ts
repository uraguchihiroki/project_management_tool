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

    await page.getByRole('button', { name: /ワークフローを追加|新規/ }).first().click()
    await page.getByPlaceholder('例: 通常承認フロー').fill(name)
    await page.getByPlaceholder('例: 一般的な業務申請に使用').fill('E2Eテスト用')
    await page.getByRole('button', { name: /追加|作成/ }).click()

    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 })
  })
})
