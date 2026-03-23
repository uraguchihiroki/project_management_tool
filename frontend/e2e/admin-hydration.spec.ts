import { test, expect } from '@playwright/test'

const API = process.env.PLAYWRIGHT_API_URL || 'http://localhost:8080/api/v1'
const DEBUG_EMAIL = 'debug_admin_1774027079@example.com'

async function getDebugUserAndOrg() {
  const userRes = await fetch(`${API}/users`)
  const userJson = await userRes.json()
  const users = userJson.data ?? []
  const user = users.find((u: any) => u.email === DEBUG_EMAIL)
  if (!user) throw new Error(`debug user not found: ${DEBUG_EMAIL}`)

  // このフロントの組織判定は /users/:id/organizations で行われる
  const orgRes = await fetch(`${API}/users/${user.id}/organizations`)
  const orgJson = await orgRes.json()
  const orgs = orgJson.data ?? orgJson ?? []
  const org = orgs[0]
  if (!org) throw new Error(`debug org not found for user=${user.id}`)

  return {
    user: { ...user, is_admin: true },
    org,
  }
}

test('admin/users: no hydration / tbody nesting errors', async ({ page }) => {
  const { user, org } = await getDebugUserAndOrg()

  const consoleErrors: string[] = []
  page.on('console', (msg) => {
    const text = msg.text()
    if (
      text.includes('hydration') ||
      text.includes('Hydration') ||
      text.includes('validateDOMNesting') ||
      text.includes('<div>') && text.includes('tbody') ||
      text.includes('tbody') && text.includes('div')
    ) {
      consoleErrors.push(text)
    }
  })

  page.on('pageerror', (err) => {
    const text = String(err)
    if (text.includes('Hydration') || text.includes('hydration') || text.includes('validateDOMNesting')) {
      consoleErrors.push(text)
    }
  })

  await page.addInitScript(
    ({ userData, orgData }: { userData: any; orgData: any }) => {
      sessionStorage.setItem('currentUser', JSON.stringify(userData))
      sessionStorage.setItem('currentOrg', JSON.stringify(orgData))
    },
    { userData: user, orgData: org }
  )

  await page.goto('/admin/users', { waitUntil: 'networkidle' })
  // Dev overlay が出ていても、少なくともDOMネスト違反/ hydration は出ないこと
  await page.waitForTimeout(1500)

  // もしエラーが出ているなら詳細を残す
  expect(consoleErrors, `console errors:\n${consoleErrors.join('\n')}`).toHaveLength(0)
})

