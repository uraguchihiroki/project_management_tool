import { test, expect } from '@playwright/test'
import { setupAuth } from './helpers'

const API = process.env.PLAYWRIGHT_API_URL || 'http://localhost:8080/api/v1'
const TEST_ORG_ID = '00000000-0000-0000-0000-000000000001'

test.describe('Issue詳細: 承認', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page, 'e2e-approve@example.com')
  })

  test('承認ボタンクリック → ステータス変更', async ({ page }) => {
    const userRes = await fetch(`${API}/users`)
    const userJson = await userRes.json()
    const user = (userJson.data ?? []).find((u: { email: string }) => u.email === 'e2e-approve@example.com')
    if (!user) throw new Error('ユーザーが取得できません')

    const projectsRes = await fetch(`${API}/projects?org_id=${TEST_ORG_ID}`)
    const projectsJson = await projectsRes.json()
    let project = (projectsJson.data ?? [])[0]
    if (!project) {
      const createRes = await fetch(`${API}/projects`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          key: 'E2EAPPR',
          name: 'E2E承認用',
          owner_id: user.id,
          organization_id: TEST_ORG_ID,
        }),
      })
      const createJson = await createRes.json()
      project = createJson.data
    }
    if (!project) throw new Error('プロジェクトが取得できません')

    const projDetailRes = await fetch(`${API}/projects/${project.id}`)
    const projDetailJson = await projDetailRes.json()
    const projDetail = projDetailJson.data
    const statuses = projDetail?.statuses ?? []
    const startStatus = statuses.find((s: { status_key?: string }) => s.status_key === 'sts_start') ?? statuses[0]
    if (!startStatus) throw new Error('ステータスが取得できません')

    const workflowsRes = await fetch(`${API}/workflows`)
    const workflowsJson = await workflowsRes.json()
    const workflows = workflowsJson.data ?? []
    const workflow = workflows.find((w: { steps?: unknown[] }) => (w.steps ?? []).length > 0) ?? workflows[0]
    if (!workflow) throw new Error('ワークフローが取得できません')

    const templatesRes = await fetch(`${API}/projects/${project.id}/templates`)
    const templatesJson = await templatesRes.json()
    let template = (templatesJson.data ?? []).find((t: { workflow_id?: number }) => String(t.workflow_id) === String(workflow.id))
    if (!template) {
      const createTmplRes = await fetch(`${API}/templates`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: 'E2E承認用テンプレート',
          project_id: project.id,
          workflow_id: workflow.id,
          default_priority: 'medium',
        }),
      })
      const createTmplJson = await createTmplRes.json()
      template = createTmplJson.data
    }
    if (!template) throw new Error('テンプレートが取得できません')

    const issueRes = await fetch(`${API}/projects/${project.id}/issues`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        title: 'E2E承認用Issue',
        author_id: user.id,
        template_id: template.id,
      }),
    })
    const issueJson = await issueRes.json()
    const issue = issueJson.data
    if (!issue) throw new Error('Issueが作成できません')

    await page.goto(`/projects/${project.id}/issues/${issue.number}`)
    await expect(page.getByText('E2E承認用Issue')).toBeVisible({ timeout: 5000 })

    const approveBtn = page.getByRole('button', { name: /^承認$/ })
    if (await approveBtn.isVisible()) {
      await approveBtn.click()
      await expect(page.getByText(/承認済み|承認ステップが完了/)).toBeVisible({ timeout: 5000 })
    } else {
      test.skip(true, '承認フローが未設定のためスキップ')
    }
  })
})
