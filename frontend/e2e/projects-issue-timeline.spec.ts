import { test, expect } from '@playwright/test'
import { setupAuth } from './helpers'

const API = process.env.PLAYWRIGHT_API_URL || 'http://localhost:8080/api/v1'
const TEST_ORG_ID = '00000000-0000-0000-0000-000000000001'

test.describe('Issue詳細: 履歴（インプリント）', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page, 'e2e-timeline@example.com')
  })

  test('履歴セクションが表示される', async ({ page }) => {
    const userRes = await fetch(`${API}/users`)
    const userJson = await userRes.json()
    const user = (userJson.data ?? []).find((u: { email: string }) => u.email === 'e2e-timeline@example.com')
    if (!user) throw new Error('ユーザーが取得できません')

    const projectsRes = await fetch(`${API}/projects?org_id=${TEST_ORG_ID}`)
    const projectsJson = await projectsRes.json()
    let project = (projectsJson.data ?? [])[0]
    if (!project) {
      const createRes = await fetch(`${API}/projects`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          key: 'E2ETL',
          name: 'E2E履歴用',
          owner_id: user.id,
          organization_id: TEST_ORG_ID,
        }),
      })
      const createJson = await createRes.json()
      project = createJson.data
    }
    if (!project) throw new Error('プロジェクトが取得できません')

    const statusesRes = await fetch(`${API}/projects/${project.id}/statuses`)
    const statusesJson = await statusesRes.json()
    const statuses = statusesJson.data ?? []
    const statusId = statuses[0]?.id
    if (!statusId) throw new Error('ステータスがありません')

    const issueRes = await fetch(`${API}/projects/${project.id}/issues`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        title: 'E2E履歴用Issue',
        status_id: statusId,
        reporter_id: user.id,
      }),
    })
    const issueJson = await issueRes.json()
    const issue = issueJson.data
    if (!issue) throw new Error('Issueが作成できません')

    await page.goto(`/projects/${project.id}/issues/${issue.number}`)
    await expect(page.getByText('E2E履歴用Issue')).toBeVisible({ timeout: 5000 })

    const timeline = page.getByTestId('issue-imprint-timeline')
    await expect(timeline).toBeVisible()
    await expect(timeline.getByRole('heading', { name: '履歴' })).toBeVisible()
  })
})
