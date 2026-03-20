import { test, expect } from '@playwright/test'
import { setupAuth } from './helpers'

const API = process.env.PLAYWRIGHT_API_URL || 'http://localhost:8080/api/v1'
const TEST_ORG_ID = '00000000-0000-0000-0000-000000000001'

test.describe('プロジェクト: Issue作成', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page, 'e2e-issues@example.com')
  })

  test('フォームで title 入力 → 作成 → カンバンに表示', async ({ page }) => {
    const unique = Date.now()
    const title = `E2E Issue ${unique}`

    const userRes = await fetch(`${API}/users`)
    const userJson = await userRes.json()
    const user = (userJson.data ?? []).find((u: { email: string }) => u.email === 'e2e-issues@example.com')
    if (!user) throw new Error('ユーザーが取得できません')

    const projectsRes = await fetch(`${API}/projects?org_id=${TEST_ORG_ID}`)
    const projectsJson = await projectsRes.json()
    let project = (projectsJson.data ?? [])[0]
    if (!project) {
      const createRes = await fetch(`${API}/projects`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          key: 'E2EISSUE',
          name: 'E2E Issue用',
          owner_id: user.id,
          organization_id: TEST_ORG_ID,
        }),
      })
      const createJson = await createRes.json()
      project = createJson.data
    }
    if (!project) throw new Error('プロジェクトが取得できません')

    await page.goto(`/projects/${project.id}`)
    await expect(page.getByRole('button', { name: /Issue作成/ })).toBeVisible({ timeout: 5000 })

    await page.getByRole('button', { name: /Issue作成/ }).click()
    await page.getByPlaceholder(/Issueのタイトルを入力/).fill(title)
    await page.locator('form').getByRole('button', { name: /作成する/ }).click()

    await expect(page.getByText(title)).toBeVisible({ timeout: 5000 })
  })
})
