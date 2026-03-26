import { test, expect } from '@playwright/test'
import { setupAuth } from './helpers'

const API = process.env.PLAYWRIGHT_API_URL || 'http://localhost:8080/api/v1'
const TEST_ORG_ID = '00000000-0000-0000-0000-000000000001'

test.describe('組織管理者: テンプレート作成', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page, 'e2e-templates@example.com')
  })

  test('プロジェクト・name 等入力 → 作成 → 一覧に表示', async ({ page }) => {
    const unique = Date.now()
    const name = `E2Eテンプレート${unique}`

    // プロジェクトを取得（なければ作成）
    const projectsRes = await fetch(`${API}/projects?org_id=${TEST_ORG_ID}`)
    const projectsJson = await projectsRes.json()
    const projects = projectsJson.data ?? []
    let projectId = projects[0]?.id
    if (!projectId) {
      const createRes = await fetch(`${API}/projects`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          key: 'E2E',
          name: 'E2Eプロジェクト',
          owner_id: (await fetch(`${API}/users`).then((r) => r.json())).data[0]?.id,
          organization_id: TEST_ORG_ID,
        }),
      })
      const createJson = await createRes.json()
      projectId = createJson.data?.id
    }
    if (!projectId) throw new Error('プロジェクトが取得できません')

    await page.goto('/admin/templates')
    await expect(page.getByRole('heading', { name: /テンプレート/ })).toBeVisible({ timeout: 5000 })

    await page.getByRole('button', { name: /テンプレートを追加/ }).first().click()
    await page.locator('select').first().selectOption(projectId)
    await page.getByPlaceholder(/例: バグ報告/).fill(name)
    await page.getByRole('button', { name: /^追加$/ }).click()

    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 })
  })
})
