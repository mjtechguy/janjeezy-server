# Admin Web UI Delivery Plan

## 1. Product Goals
- Provide a secure, admin-only Next.js 15.5 interface to configure and monitor the Jan API Gateway.
- Support both Google OAuth and local credential-based admin access, with passwords hashed via Argon2 and persisted in Postgres.
- Expose every available administrative capability offered by the Go APIs (organization, projects, providers, API keys, users, conversations, responses, MCP).
- Establish foundations (auth, data layer, UI system, testing) that scale to future endpoints like audit logging and environment controls.

## 2. High-Level Architecture
| Concern | Approach |
|---------|----------|
| Runtime | Next.js 15.5 App Router with React Server Components (RSC) default; selectively opt into Client Components for interactivity. |
| Styling | Tailwind CSS 4 (via PostCSS) + Headless UI primitives. Theme tokens stored in `app/(admin)/theme.config.ts`. |
| State/Data | TanStack Query for client caching, Zod for schema validation, Zustand for lightweight client state (e.g., session). |
| Networking | Route handlers under `app/api/jan/*/route.ts` act as a BFF, injecting admin access tokens and forwarding cookies. |
| Auth | Dual-mode auth: Google OAuth plus local username/password login. Local creds stored in Postgres with Argon2id hashing and salted per-user. Session context hydrated in `layout.tsx`. |
| Deployment | Vercel (preferred) or containerized Node 20 runtime; environment variables reference Go gateway host, OAuth secrets, JWT signing key. |

## 3. Delivery Phases & Task Breakdown

### Phase A – Foundations
1. **Bootstrap Next.js 15.5 App**
   - `npx create-next-app@latest ui-admin --ts --app --tailwind`.
   - Configure `tsconfig.json` baseUrl, path aliases (`@lib`, `@components`).
2. **Project Scaffolding**
   - Create `app/(admin)/layout.tsx` with secure layout wrapper.
   - Initialize Tailwind config for admin palette.
   - Add `eslint-config-next`, `@typescript-eslint`, `prettier` rules.
3. **Environment Management**
   - Define `.env.example` with `JAN_API_BASE_URL`, `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `NEXTAUTH_SECRET` (if using next-auth fallback).
   - Implement runtime schema using `zod` in `lib/env.ts`.

### Phase B – Authentication & Session
1. **Login Flow**
   - Admin login page at `/admin/login` offering Google OAuth and local credential panels.
   - Google path: client requests `/api/jan/auth/google/login` to fetch redirect URL, browser performs OAuth handshake, callback handled at `/admin/auth/callback`.
   - Local path: form posts email + password to `/api/jan/auth/local/login` (new API). Backend verifies Argon2 hash from Postgres `admin_users` table, returns JWT.
2. **Callback Handler**
   - Route handler posts to `/v1/auth/google/callback`, extracts access token, stores in HttpOnly cookie `jan_admin_access`.
   - Local login route handler posts to `/v1/auth/local/login`, receives access token and refresh cookie.
   - Persist refresh token through Go-set cookie; schedule refresh tasks.
3. **Session Management**
   - `useAdminSession()` hook reads cookie via server action, hydrates client store.
   - Add `RefreshControl` component using `useEffect` to call `/api/jan/auth/refresh-token` before expiry.
4. **Route Guard**
   - Middleware in `middleware.ts` blocks unauthenticated access to `/admin/(protected)` segments, redirecting to login.

#### Example: Route Handler for Login URL
```ts
// app/api/jan/auth/google/login/route.ts
import { NextResponse } from 'next/server';
import { janFetch } from '@/lib/jan-fetch';

export async function GET() {
  const res = await janFetch('/v1/auth/google/login', { method: 'GET' });
  if (!res.ok) {
    return NextResponse.json({ error: 'Unable to init login' }, { status: 500 });
  }
  const body = await res.json();
  return NextResponse.json({ redirectUrl: body.url });
}
```

#### Example: Local Login Route
```ts
// app/api/jan/auth/local/login/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { janFetch } from '@/lib/jan-fetch';

export async function POST(req: NextRequest) {
  const credentials = await req.json();
  const res = await janFetch('/v1/auth/local/login', {
    method: 'POST',
    body: JSON.stringify(credentials),
    headers: { 'Content-Type': 'application/json' },
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'Invalid credentials' }));
    return NextResponse.json(err, { status: res.status });
  }

  const body = await res.json();
  const response = NextResponse.json(body);
  copySetCookiesFromUpstream(res, response); // helper to forward refresh token cookie
  return response;
}
```

### Phase C – Networking & Data Utilities
1. **Jan Fetch Wrapper**
   - `lib/jan-fetch.ts` adds base URL, attaches `Authorization: Bearer <access>` when available, and forwards cookies on server.
   - Handles 401 -> triggers re-auth redirect.
2. **API Client Modules**
   - `services/organization.ts`, `services/projects.ts`, etc., containing strongly typed wrappers using Zod parsing.
3. **Error & Toast System**
   - Global toaster component (e.g., `@radix-ui/react-toast`).
   - `handleApiError(err)` helper mapping server codes to UI copy.

#### Example: Shared Fetch Wrapper
```ts
export async function janFetch(path: string, init?: RequestInit) {
  const url = new URL(path, process.env.JAN_API_BASE_URL).toString();
  const headers = new Headers(init?.headers);

  const token = await getServerToken(); // reads secure cookie / session store
  if (token) headers.set('Authorization', `Bearer ${token}`);

  const response = await fetch(url, { ...init, headers, credentials: 'include' });
  if (response.status === 401) throw new UnauthorizedError();
  return response;
}
```

### Phase D – Core Layout & Navigation
1. **Shell**
   - Sidebar navigation with sections: Overview, Organization, Projects, Providers, Users, API Keys, Conversations, Responses, MCP.
   - `Breadcrumbs` derived from current segment metadata.
   - `TopBar` showing admin identity (`/v1/auth/me`) and quick actions.
2. **Dashboard**
   - Cards for total projects, active providers, guest users, recent invites.
   - Health widget hitting `/healthcheck` and `/v1/version`.

### Phase E – Feature Modules

#### 1. Organization & Project Admin
Tasks:
- List projects (paginated using `/v1/organization/projects?limit`).
- Create/rename/archive project via POST routes.
- Manage default provider metadata (patch metadata payload).
- View members per project (tie into invites & membership endpoints).

Key UI components:
- `ProjectTable` with infinite scroll.
- `ProjectDrawer` showing details and action buttons (archive, register provider).

#### 2. Provider Management
Tasks:
- Global providers page using `/v1/models/providers`.
- Group by scope (jan / organization / project).
- Register new provider to project: POST `/v1/organization/projects/:id/models/providers`.
- Update provider: PATCH with active toggle, metadata edit, key rotation prompt.

Code snippet for server action:
```ts
'use server';
import { janFetch } from '@/lib/jan-fetch';

export async function registerProjectProvider(projectId: string, payload: ProviderFormData) {
  const res = await janFetch(`/v1/organization/projects/${projectId}/models/providers`, {
    method: 'POST',
    body: JSON.stringify(payload),
    headers: { 'Content-Type': 'application/json' },
  });
  if (!res.ok) throw await mapJanError(res);
  revalidatePath(`/admin/projects/${projectId}`);
}
```

#### 3. User & Invite Management
Tasks:
- List org members with roles (requires backend invite/member endpoints).
- Issue invites (POST `/v1/organization/invites`).
- Resend or revoke invite.
- Convert guest users to full accounts (trigger Google connect path).

Components:
- `MemberList` with search, role badges.
- `InviteForm` modal supporting bulk email entry.

#### 4. API Key Administration
Tasks:
- Admin API keys: list, create, revoke (existing endpoints).
- Project API keys: nested under project detail with single-view download.
- Track last-used timestamps when provided by backend.

UX:
- Generate modal shows token once, encourages secure storage.
- Copy-to-clipboard with confirmation banner.

#### 5. Conversation & Workspace Oversight
Tasks:
- Paginated conversation list (`/v1/conversations` with workspace filters).
- View single conversation, list items (`/items` endpoint), delete conversation.
- Move conversation between workspaces via PATCH `/workspace`.
- Display metadata and stored reasoning text for debugging.

#### 6. Responses Monitor
Tasks:
- Trigger demonstration responses via `/v1/responses`.
- Poll status (if background) and cancel as needed.
- Render output blocks (text, annotations, files).

#### 7. MCP Diagnostics
Tasks:
- Connect to `/v1/mcp` using EventSource.
- List tools, prompts, resources; execute test queries.
- Show raw JSON exchanged for troubleshooting.

### Phase F – Cross-Cutting Enhancements
1. **Activity/Audit Feed** (placeholder until backend exposes endpoint).
2. **Role-Based Guards**: restrict destructive actions to org owners.
3. **Internationalization**: set up next-intl for future translations.
4. **Accessibility**: ensure keyboard and screen-reader coverage.

### Phase G – Testing & Quality
1. **Unit Tests**: Vitest + React Testing Library for hooks/components.
2. **Integration Tests**: Mock service modules verifying API contracts.
3. **E2E Tests**: Playwright scripts for login, project creation, provider registration, API key generation, conversation review.
4. **CI Pipeline**: GitHub Actions running lint, type check, test, Playwright.
5. **Security**: Add `helmet`-equivalent headers via Next middleware; enforce CSP referencing gateway domain.

### Phase H – Deployment & Ops
1. **Preview Environments**: per-branch Vercel preview hitting staging API gateway.
2. **Config Promotion**: Document env variable promotion path (dev → staging → prod).
3. **Monitoring**: Hook Vercel analytics + Sentry for client errors; tie into existing Grafana dashboards for API logs.

## 4. Detailed Task Matrix

| Epic | Task | Est. |
|------|------|------|
| Foundations | Scaffold Next.js app, Tailwind, linting | 2d |
| Auth | Implement Google OAuth + local Argon2 login flows, session store, guard | 4d |
| Networking | Build janFetch, service modules, error handling | 2d |
| Layout | Sidebar/topbar/dash, health widgets | 3d |
| Organization | Project list CRUD, metadata editing | 4d |
| Providers | Catalog, register/update forms | 4d |
| Users/Invites | Member list, invite management | 3d |
| API Keys | Admin + project key CRUD UX | 3d |
| Conversations | List/detail/workspace actions | 5d |
| Responses | Monitor, cancel, output rendering | 3d |
| MCP | Diagnostic console with streaming | 4d |
| Cross-cutting | Role guards, a11y, i18n baseline | 3d |
| Testing/CI | Vitest, Playwright, GitHub Actions | 4d |
| Docs/Runbooks | Admin handbook, environment guide | 2d |

_Total_: ~41 engineer-days (single resource). Parallelization across epics will reduce calendar time.

## 5. Component Inventory
- `components/shell/Sidebar.tsx`
- `components/shell/TopBar.tsx`
- `components/data/DataTable.tsx` (generic table with pagination).
- `components/forms/JsonMetadataEditor.tsx`.
- `components/providers/ProviderStatusBadge.tsx`.
- `components/conversations/ConversationViewer.tsx`.
- `components/mcp/McpConsole.tsx`.

All components documented in Storybook for design QA.

## 6. Data Contracts & Zod Schemas
Create schema files per resource in `schemas/`:
```ts
export const ProviderSchema = z.object({
  id: z.string(),
  slug: z.string(),
  name: z.string(),
  vendor: z.string(),
  base_url: z.string().url().nullable(),
  active: z.boolean(),
  metadata: z.record(z.string()).optional(),
  scope: z.enum(['jan', 'organization', 'project']),
  project_id: z.string().nullable(),
});
export type Provider = z.infer<typeof ProviderSchema>;

export const LocalLoginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(12, 'Password must be at least 12 characters'),
});
```

## 7. Documentation & Runbooks
- `docs/admin/architecture.md` — diagram of auth/token flow.
- `docs/admin/deployment.md` — environment setup, CI/CD steps.
- `docs/admin/operations.md` — how to recover locked sessions, rotate provider credentials, manage local login lifecycle (create user, reset password), monitor MCP.
- Update main `README` with link to admin UI instructions.

## 8. Open Questions / Dependencies
1. Confirm availability of invite/member endpoints and their exact routes.
2. Determine if backend emits audit logs; if not, plan placeholder UI.
3. Clarify requirement for guest account conversion workflows.
4. Verify SSO beyond Google (Azure AD/Okta?) for future expansion.
5. Define backend plan for seeding/maintaining local admin users in Postgres, including password reset flows and Argon2 parameters (memory, iterations, parallelism).

## 9. Next Actions
- [ ] Validate backend API coverage against admin needs; schedule backend enhancements for gaps (audit logs, env flags).
- [ ] Approve UX wireframes for navigation and key workflows.
- [ ] Kick off Phase A tasks and set up CI pipeline.
- [ ] Coordinate with backend team to fix identified issues (user pagination, refresh token duration, multi-modal storage) before UI depends on them.
