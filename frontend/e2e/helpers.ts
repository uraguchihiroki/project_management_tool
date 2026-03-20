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

/**
 * `ensureLoginableUser` と同じく、任意メールでログイン可能なユーザーを fetch で保証する。
 * （setupAuth は Node 上の fetch しか使えないため）
 */
async function ensureLoginableUserFetch(email: string): Promise<void> {
  const login = await fetch(`${API}/admin/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email }),
  })
  if (login.ok) return

  const tryCreateUser = () =>
    fetch(`${API}/users`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: 'E2E Playwright Login', email }),
    })

  let create = await tryCreateUser()
  if (create.ok) return

  let body = await create.text()
  const needsOrg =
    create.status === 500 &&
    (body.includes('organization') ||
      body.includes('組織') ||
      body.includes('組織が存在') ||
      body.includes('23502') ||
      body.includes('null value'))

  if (needsOrg) {
    const saLogin = await fetch(`${API}/super-admin/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: E2E_SUPER_ADMIN_EMAIL }),
    })
    if (!saLogin.ok) {
      const t = await saLogin.text()
      throw new Error(
        `ensureLoginableUserFetch: 組織があり POST /users できません。super-admin ログイン失敗 ${saLogin.status} ${t}`,
      )
    }
    const saBody = (await saLogin.json()) as { token?: string }
    const token = saBody.token
    if (!token) {
      throw new Error('ensureLoginableUserFetch: super-admin/login に token がありません')
    }
    const orgsRes = await fetch(`${API}/super-admin/organizations`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    if (!orgsRes.ok) {
      const t = await orgsRes.text()
      throw new Error(`ensureLoginableUserFetch: GET /super-admin/organizations failed ${orgsRes.status} ${t}`)
    }
    const orgsJson = (await orgsRes.json()) as { data?: unknown[] }
    const orgs = orgsJson.data ?? []
    if (orgs.length === 0) {
      const createOrg = await fetch(`${API}/super-admin/organizations`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: 'E2E Bootstrap Organization' }),
      })
      if (!createOrg.ok) {
        const t = await createOrg.text()
        throw new Error(`ensureLoginableUserFetch: POST /super-admin/organizations failed ${createOrg.status} ${t}`)
      }
    }
    create = await tryCreateUser()
    if (create.ok) return
    body = await create.text()
  }

  throw new Error(
    `ensureLoginableUserFetch: POST /users failed ${create.status} ${body}\n` +
      `（ヒント: backend/seed.sql を投入し、バックエンドを JWT 対応ビルドにしてください。）`,
  )
}

export async function setupAuth(page: import('@playwright/test').Page, email = 'e2e-create@example.com') {
  // GET /users は JWT 必須のため未認証では一覧できない。未登録なら POST /users で作成する。
  await ensureLoginableUserFetch(email)

  const loginRes = await fetch(`${API}/admin/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email }),
  })
  const loginJson = (await loginRes.json()) as {
    data?: { token?: string; user?: Record<string, unknown> }
  }
  if (!loginRes.ok) {
    throw new Error(
      `setupAuth: POST /admin/login failed ${loginRes.status} ${JSON.stringify(loginJson)}`,
    )
  }
  let token = loginJson.data?.token
  let userForSession = loginJson.data?.user
  if (!token || !userForSession) {
    throw new Error('setupAuth: admin/login に token または user がありません')
  }

  const uid = String(userForSession.id ?? '')
  if (!uid) {
    throw new Error('setupAuth: admin/login の user に id がありません')
  }

  const userOrgsRes = await fetch(`${API}/users/${uid}/organizations`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  const userOrgsJson = (await userOrgsRes.json()) as { data?: Array<{ id: string; name: string }> }
  if (!userOrgsRes.ok) {
    throw new Error(
      `setupAuth: GET /users/${uid}/organizations failed ${userOrgsRes.status} ${JSON.stringify(userOrgsJson)}`,
    )
  }
  const userOrgs = userOrgsJson.data ?? []
  if (userOrgs.length === 0) {
    throw new Error(
      `setupAuth: ユーザー ${email} の所属組織がありません。seed または super-admin で組織を作成してください。`,
    )
  }

  const loginOrgId = String(userForSession.organization_id ?? '')
  // テスト用 FRS を優先。switch が無い古いバックエンドでは後段で JWT の組織に矯正する。
  let orgForSession =
    userOrgs.find((o) => o.id === TEST_ORG_ID) ??
    userOrgs.find((o) => o.id === loginOrgId) ??
    userOrgs[0]

  const switchRes = await fetch(`${API}/admin/switch-organization`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
    body: JSON.stringify({ organization_id: orgForSession.id }),
  })
  const switchJson = (await switchRes.json()) as {
    data?: { token?: string; user?: Record<string, unknown> }
  }
  if (switchRes.ok) {
    token = switchJson.data?.token ?? token
    userForSession = switchJson.data?.user ?? userForSession
  } else if (switchRes.status === 404) {
    // ルート未実装: session の currentOrg を JWT の organization_id と一致させないと API と UI がズレる
    const aligned = userOrgs.find((o) => o.id === loginOrgId) ?? userOrgs[0]
    if (!aligned) {
      throw new Error('setupAuth: switch-organization なしで組織を JWT に合わせられません')
    }
    orgForSession = aligned
  } else {
    throw new Error(
      `setupAuth: POST /admin/switch-organization failed ${switchRes.status} ${JSON.stringify(switchJson)}`,
    )
  }

  await page.addInitScript(
    ({
      authToken,
      userData,
      orgData,
    }: {
      authToken: string
      userData: Record<string, unknown>
      orgData: { id: string; name: string }
    }) => {
      sessionStorage.setItem('authToken', authToken)
      sessionStorage.setItem('currentUser', JSON.stringify({ ...userData, is_admin: true }))
      sessionStorage.setItem('currentOrg', JSON.stringify(orgData))
    },
    { authToken: token, userData: userForSession, orgData: orgForSession },
  )
  return { user: userForSession, org: orgForSession }
}
