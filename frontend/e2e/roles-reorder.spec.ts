import { test, expect } from '@playwright/test'

const API = process.env.PLAYWRIGHT_API_URL || 'http://localhost:8080/api/v1'
const TEST_ORG_ID = '00000000-0000-0000-0000-000000000001' // FRS from seed

test.describe('役職管理 ドラッグ並び替え', () => {
  test.beforeEach(async ({ page }) => {
    // テスト用ユーザー・組織をAPIで準備し、sessionStorage にセット
    const userRes = await fetch(`${API}/users`)
    const userJson = await userRes.json()
    const users = userJson.data ?? []
    let user = users.find((u: { email: string }) => u.email === 'e2e-role@example.com')
    if (!user) {
      const createRes = await fetch(`${API}/users`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: 'E2E役職', email: 'e2e-role@example.com' }),
      })
      const createJson = await createRes.json()
      user = createJson.data
    }
    await fetch(`${API}/users/${user.id}/admin`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ is_admin: true }),
    })
    const orgRes = await fetch(`${API}/organizations`)
    const orgJson = await orgRes.json()
    const orgs = orgJson.data ?? []
    const org = orgs.find((o: { id: string }) => o.id === TEST_ORG_ID) ?? orgs[0] ?? { id: TEST_ORG_ID, name: 'FRS' }
    await fetch(`${API}/organizations/${org.id}/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: user.id, is_org_admin: true }),
    }).catch(() => {})

    await page.addInitScript(
      ({ userData, orgData }: { userData: { id: string; name: string; email: string }; orgData: { id: string; name: string } }) => {
        sessionStorage.setItem('currentUser', JSON.stringify({ ...userData, is_admin: true }))
        sessionStorage.setItem('currentOrg', JSON.stringify(orgData))
      },
      { userData: user, orgData: org }
    )
  })

  test('reorder API で役職の並び順を変更できる', async ({ page }) => {
    // テスト用に2件作成して並び替えを検証
    await fetch(`${API}/roles`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: 'E2E役職X', level: 1, organization_id: TEST_ORG_ID }),
    })
    await fetch(`${API}/roles`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: 'E2E役職Y', level: 2, organization_id: TEST_ORG_ID }),
    })
    const rolesRes = await fetch(`${API}/roles?org_id=${TEST_ORG_ID}`)
    const roles = (await rolesRes.json()).data ?? []
    const ids = roles.filter((r: { name: string }) => r.name.startsWith('E2E役職')).map((r: { id: number }) => r.id)
    if (ids.length < 2) throw new Error('役職が2件以上必要です')
    const reordered = [ids[1], ids[0]]
    const reorderRes = await fetch(`${API}/roles/bulk/reorder?org_id=${TEST_ORG_ID}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ids: reordered }),
    })
    expect(reorderRes.status, 'reorder API が 204 で成功すること').toBe(204)

    await page.goto('/admin/roles')
    const rows = page.locator('tbody tr')
    await expect(rows.first()).toBeVisible({ timeout: 5000 })
    // 並び替え後: 元2番目(E2E役職Y)が1番目、元1番目(E2E役職X)が2番目
    const firstRow = rows.filter({ hasText: 'E2E役職' }).first()
    const secondRow = rows.filter({ hasText: 'E2E役職' }).nth(1)
    await expect(firstRow.locator('td:nth-child(2)')).toHaveText('E2E役職Y', { timeout: 3000 })
    await expect(secondRow.locator('td:nth-child(2)')).toHaveText('E2E役職X')
  })

  test('ドラッグで役職の並び順を変更できる', async ({ page }) => {
    // reorder API の呼び出しとレスポンスを検証
    let reorderStatus: number | null = null
    page.on('response', async (res) => {
      if (res.url().includes('/roles/bulk/reorder') && res.request().method() === 'PUT') {
        reorderStatus = res.status()
      }
    })

    await page.goto('/admin/roles')

    // 役職が2件以上ない場合はAPIで作成
    const rolesRes = await fetch(`${API}/roles?org_id=${TEST_ORG_ID}`)
    const rolesJson = await rolesRes.json()
    const roles = rolesJson.data ?? []
    if (roles.length < 2) {
      await fetch(`${API}/roles`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: '役職A',
          level: 5,
          organization_id: TEST_ORG_ID,
        }),
      })
      await fetch(`${API}/roles`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: '役職B',
          level: 7,
          organization_id: TEST_ORG_ID,
        }),
      })
      await page.reload()
    }

    const rows = page.locator('tbody tr')
    await expect(rows.first()).toBeVisible({ timeout: 5000 })
    expect(await rows.count()).toBeGreaterThanOrEqual(2)

    const firstRow = rows.first()
    const secondRow = rows.nth(1)
    const firstName = await firstRow.locator('td:nth-child(2)').textContent()
    const secondName = await secondRow.locator('td:nth-child(2)').textContent()

    // 1行目を2行目の位置にドラッグ（steps で中間 mousemove を発生させ dnd-kit の activation を満たす）
    const grip = firstRow.locator('[title="ドラッグして並び替え"]').first()
    const reorderPromise = page.waitForResponse((r) => r.url().includes('/roles/bulk/reorder') && r.request().method() === 'PUT', { timeout: 5000 })
    await grip.dragTo(secondRow, { steps: 20 })
    const reorderRes = await reorderPromise
    if (reorderRes.status() !== 204) {
      const body = await reorderRes.text()
      throw new Error(`reorder API: ${reorderRes.status()} body=${body}`)
    }

    // 並び替え後: 元の2行目が1行目、元の1行目が2行目になっている
    await expect(rows.first().locator('td:nth-child(2)')).toHaveText(secondName!, { timeout: 3000 })
    await expect(rows.nth(1).locator('td:nth-child(2)')).toHaveText(firstName!)
  })
})
