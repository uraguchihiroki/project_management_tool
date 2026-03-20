const API = process.env.PLAYWRIGHT_API_URL || 'http://localhost:8080/api/v1'
export const TEST_ORG_ID = '00000000-0000-0000-0000-000000000001'

export async function setupAuth(page: import('@playwright/test').Page, email = 'e2e-create@example.com') {
  const userRes = await fetch(`${API}/users`)
  const userJson = await userRes.json()
  const users = userJson.data ?? []
  let user = users.find((u: { email: string }) => u.email === email)
  if (!user) {
    const createRes = await fetch(`${API}/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: 'E2E作成', email }),
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
  return { user, org }
}
