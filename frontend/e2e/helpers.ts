import type { APIRequestContext } from '@playwright/test'

const API = process.env.PLAYWRIGHT_API_URL || 'http://localhost:8080/api/v1'
export const TEST_ORG_ID = '00000000-0000-0000-0000-000000000001'

/** ログイン E2E 専用。`ensureLoginableUser` で未登録なら POST /users する。 */
export const E2E_LOGIN_EMAIL = 'e2e-playwright-login@example.com'

/** seed.sql のスーパー管理者（組織0件時に POST /super-admin/organizations でブートストラップ） */
const E2E_SUPER_ADMIN_EMAIL = process.env.E2E_SUPER_ADMIN_EMAIL || 'superadmin@frs.example.com'

/**
 * 組織が0件のとき、スーパー管理者でログインして組織を1件作成する（seed に super_admins がある前提）。
 */
async function ensureOrganizationExistsViaSuperAdmin(request: APIRequestContext): Promise<void> {
  const login = await request.post(`${API}/super-admin/login`, {
    data: JSON.stringify({ email: E2E_SUPER_ADMIN_EMAIL }),
    headers: { 'Content-Type': 'application/json' },
  })
  if (!login.ok()) {
    const t = await login.text()
    throw new Error(
      `ensureLoginableUser: 組織があり POST /users できません。super-admin ログイン失敗 ${login.status()} ${t}\n` +
        `→ backend/seed.sql を実行するか、E2E_SUPER_ADMIN_EMAIL をシード済みのスーパー管理者に合わせてください。`,
    )
  }
  const body = (await login.json()) as { token?: string }
  const token = body.token
  if (!token) {
    throw new Error(
      'ensureLoginableUser: super-admin/login に token がありません。バックエンドを最新ビルドにしてください。',
    )
  }

  const orgsRes = await request.get(`${API}/super-admin/organizations`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!orgsRes.ok()) {
    const t = await orgsRes.text()
    throw new Error(`ensureLoginableUser: GET /super-admin/organizations failed ${orgsRes.status()} ${t}`)
  }
  const orgsJson = (await orgsRes.json()) as { data?: unknown[] }
  const orgs = orgsJson.data ?? []
  if (orgs.length > 0) return

  const createOrg = await request.post(`${API}/super-admin/organizations`, {
    headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
    data: JSON.stringify({ name: 'E2E Bootstrap Organization' }),
  })
  if (!createOrg.ok()) {
    const t = await createOrg.text()
    throw new Error(`ensureLoginableUser: POST /super-admin/organizations failed ${createOrg.status()} ${t}`)
  }
}

/**
 * POST /admin/login が成功するユーザーを保証する。
 * 未登録なら `POST /users`（公開 API）。組織が無い場合はスーパー管理者で組織を1件作成してから再試行。
 */
export async function ensureLoginableUser(request: APIRequestContext): Promise<void> {
  const login = await request.post(`${API}/admin/login`, {
    data: JSON.stringify({ email: E2E_LOGIN_EMAIL }),
    headers: { 'Content-Type': 'application/json' },
  })
  if (login.ok()) return

  const tryCreateUser = () =>
    request.post(`${API}/users`, {
      data: JSON.stringify({ name: 'E2E Playwright Login', email: E2E_LOGIN_EMAIL }),
      headers: { 'Content-Type': 'application/json' },
    })

  let create = await tryCreateUser()
  if (create.ok()) return

  let body = await create.text()
  const needsOrg =
    create.status() === 500 &&
    (body.includes('organization') ||
      body.includes('組織') ||
      body.includes('組織が存在') ||
      body.includes('23502') ||
      body.includes('null value'))

  if (needsOrg) {
    await ensureOrganizationExistsViaSuperAdmin(request)
    create = await tryCreateUser()
    if (create.ok()) return
    body = await create.text()
  }

  throw new Error(
    `ensureLoginableUser: POST /users failed ${create.status()} ${body}\n` +
      `（ヒント: backend/seed.sql を投入し、バックエンドを JWT 対応ビルドにしてください。）`,
  )
}

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
