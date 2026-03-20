import { test, expect } from '@playwright/test'
import { setupAuth } from './helpers'

const API = process.env.PLAYWRIGHT_API_URL || 'http://localhost:8080/api/v1'
const TEST_ORG_ID = '00000000-0000-0000-0000-000000000001'

test.describe('組織管理者: ワークフロー ステップ追加', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuth(page, 'e2e-wf-step@example.com')
  })

  test('ステータス選択 → ステップ追加 → 一覧に表示', async ({ page }) => {
    const statusesRes = await fetch(`${API}/organizations/${TEST_ORG_ID}/statuses?type=issue&exclude_system=1`)
    const statusesJson = await statusesRes.json()
    const statuses = statusesJson.data ?? []
    const firstStatus = statuses[0]
    if (!firstStatus) throw new Error('ステータスが取得できません（ユーザー作成ステータスが必要）')

    const workflowsRes = await fetch(`${API}/workflows`)
    const workflowsJson = await workflowsRes.json()
    let workflow = (workflowsJson.data ?? [])[0]
    if (!workflow) {
      const createRes = await fetch(`${API}/workflows`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          organization_id: TEST_ORG_ID,
          name: 'E2Eワークフロー',
          description: '',
        }),
      })
      const createJson = await createRes.json()
      workflow = createJson.data
    }
    if (!workflow) throw new Error('ワークフローが取得できません')

    await page.goto(`/admin/workflows/${workflow.id}`)
    await expect(page.getByText(workflow.name)).toBeVisible({ timeout: 5000 })

    await page.getByRole('button', { name: /ステップを追加/ }).click()
    await page.locator('select').first().selectOption(firstStatus.id)
    await page.locator('form').getByRole('button', { name: /^追加$/ }).click()

    await expect(page.getByText(firstStatus.name)).toBeVisible({ timeout: 5000 })
  })
})
