import { test, expect } from '@playwright/test'
import { setupAuth } from './helpers'

const API = process.env.PLAYWRIGHT_API_URL || 'http://localhost:8080/api/v1'
const TEST_ORG_ID = '00000000-0000-0000-0000-000000000001'

test.describe('Issue詳細: コメント投稿', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page, 'e2e-comment@example.com')
  })

  test('テキスト入力 → 投稿 → コメント一覧に表示', async ({ page }) => {
    const unique = Date.now()
    const commentText = `E2Eコメント ${unique}`

    const userRes = await fetch(`${API}/users`)
    const userJson = await userRes.json()
    const user = (userJson.data ?? []).find((u: { email: string }) => u.email === 'e2e-comment@example.com')
    if (!user) throw new Error('ユーザーが取得できません')

    const projectsRes = await fetch(`${API}/projects?org_id=${TEST_ORG_ID}`)
    const projectsJson = await projectsRes.json()
    let project = (projectsJson.data ?? [])[0]
    if (!project) {
      const createRes = await fetch(`${API}/projects`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          key: 'E2ECOM',
          name: 'E2Eコメント用',
          owner_id: user.id,
          organization_id: TEST_ORG_ID,
        }),
      })
      const createJson = await createRes.json()
      project = createJson.data
    }
    if (!project) throw new Error('プロジェクトが取得できません')

    const issueRes = await fetch(`${API}/projects/${project.id}/issues`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        title: 'E2Eコメント用Issue',
        author_id: user.id,
      }),
    })
    const issueJson = await issueRes.json()
    const issue = issueJson.data
    if (!issue) throw new Error('Issueが作成できません')

    await page.goto(`/projects/${project.id}/issues/${issue.number}`)
    await expect(page.getByText('E2Eコメント用Issue')).toBeVisible({ timeout: 5000 })

    const commentInput = page.getByPlaceholder(/コメントを入力/)
    await commentInput.fill(commentText)
    await commentInput.locator('..').getByRole('button').click()

    await expect(page.getByText(commentText)).toBeVisible({ timeout: 5000 })
  })
})
