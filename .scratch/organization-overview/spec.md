# Organization overview dashboard (Masbro Dashboard v2)

Status: ready-for-agent

Design source: Claude Design project `2b05f82e-f1ff-442a-9213-77179cb82ffd`, file `Masbro Dashboard v2.dc.html`, the org-level screen (`view: 'overview'`).

## Problem Statement

A signed-in user lands on a placeholder page that says "Signed in as …" and nothing else. They cannot see which Organization they are in or which Projects it has, and they have no sense of where things stand. The team has a finished design for the org-level dashboard (header, org sidebar, Overview with project cards), but the frontend has no layout, no navigation and no styling that matches it. Auth work is still in progress, so the dashboard cannot rely on a real session yet.

## Solution

Replace the placeholder landing page with the org-level dashboard from the v2 design:

- **Header:** the Masbro Cloud mark, a breadcrumb showing the current Organization, a search box (visual only), Docs and Feedback links, and the user's initials avatar.
- **Org sidebar:** an org card with the name and member count, nav links for Overview, Members, Git connections and Organization settings, a Projects list with app counts and a "+" button, and a node status footer.
- **Overview main area:** the org name as the page title, a "New Project" button, a grid of Project cards and a dashed "Create a project" card.

Auth is bypassed for now. A `useAuth` hook returns a mocked user whose shape matches the backend's `/auth/me` contract. The dashboard's organization and project data also comes from hooks that return fixtures shaped like the backend contracts. When real auth lands, each hook's body is swapped for a call through `createApi()` and the components do not change.

Only the screens the design actually draws are in scope. The design has nav entries (Members, Git connections, Organization settings) and actions (New Project, search) with no screen behind them. These are rendered but do nothing.

## User Stories

1. As a signed-in user, I want to land on an org Overview instead of a placeholder page, so that I immediately see my Organization's Projects.
2. As a user, I want the Masbro Cloud mark in the top-left of every dashboard page, so that I always know which product I'm in.
3. As a user, I want the breadcrumb to show my current Organization's name after the logo, so that I know which Organization I'm looking at.
4. As a user, I want to click the Organization name in the breadcrumb to return to the Overview, so that I have a consistent way home.
5. As a user, I want a search box with a ⌘K hint in the header, so that I know search will be available (it is visual only for now).
6. As a user, I want Docs and Feedback links in the header, so that I know where help and feedback will live.
7. As a user, I want an avatar with my initials in the header, so that I can confirm which account I'm using.
8. As a user, I want my initials taken from my full name (first letters of the first two words, uppercased), so that the avatar matches me.
9. As a user whose full name is a single word, I want the avatar to show one or two letters from it, so that the avatar never renders empty.
10. As a user, I want an org card at the top of the sidebar with the Organization name and "Organization · N members", so that I see the Organization's size at a glance.
11. As a user, I want the org card to show a switcher affordance (⇅), so that I know switching Organizations will live there (not functional yet).
12. As a user, I want an Overview nav item that is highlighted while I'm on the Overview, so that I know where I am.
13. As a user, I want a Members nav item showing the member count, so that I can see team size before the Members page exists.
14. As a user, I want Git connections and Organization settings nav items, so that I can see where those features will live (not functional yet).
15. As a user, I want a Projects section in the sidebar listing every Project in the current Organization with its app count, so that I can scan all Projects without leaving the page.
16. As a user, I want a "+" button next to the Projects heading, so that I know where to create a Project (not functional yet).
17. As a user, I want a footer at the bottom of the sidebar with a node health dot and "node · X of Y GB used", so that I get a quick read on my host's capacity.
18. As a user, I want the Overview page title to be my Organization's name with an "Overview" eyebrow above it, so that the page is clearly labelled.
19. As a user, I want a primary "New Project" button on the Overview, so that the main action is obvious (not functional yet).
20. As a user, I want each Project shown as a card with its name and environment badge (e.g. production, staging), so that I can tell Projects apart.
21. As a user, I want each Project card to list its Applications as chips with a status dot (green running, amber building, red failed), so that I can spot broken apps at a glance.
22. As a user, I want each Project card's footer to show the app count, the last deploy outcome with relative time (e.g. "deployed 12m ago", "failed 1h ago") and the repo count, so that I can see recent activity without opening the Project.
23. As a user, I want "1 app" / "1 repo" to be singular and other counts plural, so that the copy reads naturally.
24. As a user, I want a dashed "Create a project" card at the end of the grid, so that there is always an entry point to add a Project.
25. As a user, I want the project grid to reflow into as many columns as fit (minimum card width 300px), so that the page uses wide screens well.
26. As a user with no Projects, I want the grid to show only the "Create a project" card and the sidebar Projects list to be empty, so that the empty state still guides me.
27. As a user, I want cards and nav items to show the hover states from the design, so that the UI feels responsive.
28. As a user, I want the dashboard in the design's dark palette with Space Grotesk for UI text and IBM Plex Mono for counts and code-like labels, so that it matches the approved design.
29. As a user, I want the header fixed at the top and the sidebar and main area to scroll on their own, so that navigation stays visible on long pages.
30. As a user, I want the last Organization I viewed remembered in this browser, so that I return to it next time.
31. As a user in several Organizations, I want the dashboard to fall back to my first Organization when there is no remembered one or the remembered one is no longer in my list, so that I never land on a broken page.
32. As a user with no Organizations, I want a simple empty state in the main area instead of a crash, so that the app degrades gracefully.
33. As a developer, I want a `useAuth` hook that returns the current user in the shape the backend's `/auth/me` returns (mapped to camelCase `User`), so that components never depend on how auth is resolved.
34. As a developer, I want `useAuth` to return a mocked user for now, so that I can build and view the dashboard without a working session.
35. As a developer, I want the organization and project data behind hooks that return fixtures shaped like the backend's list responses, so that switching to real API calls later is a change inside the hooks only.
36. As a developer, I want the dashboard reachable without signing in while auth is bypassed, so that I can open it directly in `pnpm dev`.
37. As a developer, I want the existing signin/signup pages and the `createApi()` auth client left as they are, so that the auth work in progress isn't disturbed.
38. As a developer, I want the accent color defined once as a theme token (teal `#5FB8A5`, the design default), so that it's easy to change later.
39. As a developer, I want the placeholder "Signed in as …" landing page removed, so that the dashboard is the only landing page and there is no dead code.
40. As a developer, I want the unused Supabase JS SDK dependency removed, so that the frontend doesn't ship or install a client the backend-owned auth never uses.
41. As a developer, I want the unused `VITE_SUPABASE_*` variables removed from the frontend env, so that nobody thinks the frontend talks to Supabase directly.
42. As a developer, I want the bootstrap global styles (the light/dark `color-scheme` and the `system-ui` root font) replaced by the design's dark theme and fonts, so that the global styles don't fight the dashboard styles.

## Implementation Decisions

**Scope**

- Only the org-level Overview screen of the v2 design is built. Project detail (the Applications canvas and project Members tab) is not built. The Members, Git connections and Organization settings nav items, the "+" / "New Project" / "Create a project" actions, search, Docs, Feedback and the org switcher are rendered as inert UI with no routes behind them.
- This is frontend-only. The backend is not changed.

**Auth bypass**

- Add a `useAuth` hook to the frontend. For now it returns a hard-coded mocked `User`. The `User` type is the existing one from the api module (`userId`, `email`, `fullName`), which mirrors the backend's `AuthMeResponse` (`user_id`, `email`, `full_name`). The mock uses those fields only; no fields the backend doesn't return are invented.
- The hook's return shape should hold up when it later becomes real: at least `{ user }`, where `user` is a `User`. When auth is wired, the hook will read `sessionQuery` and the route guard will come back. Components call only `useAuth`.
- The dashboard route must not redirect to `/signin` while the bypass is active. Either the Overview route moves out from under the `_authed` layout, or the `_authed` guard is bypassed. Pick whichever keeps the later revert smallest, and mark the bypass point with a short docstring note saying it is temporary.
- The signin/signup routes, `sessionQuery` and `createApi().auth` are unchanged.

**Data hooks and contracts**

- Organizations: a hook returns the user's Organizations in the shape of the backend's `GET /api/organizations` response (`ListMyOrgEntry`: `role` + `organization { org_id, name, created_at, updated_at }`), mapped to camelCase the same way `auth.ts` maps its responses.
- Projects: a hook returns the Projects for the current Organization. The backend's `GET /api/projects` returns `ListMyProjectEntry` (`role` + `project { org_id, project_id, name, created_at, updated_at }`) for all of the user's projects, so the hook filters by `org_id`.
- The design shows fields the backend does not provide yet: the org member count, a project's environment, per-project Applications with status, last deploy outcome and time, repo count, and node capacity. These are added to the fixtures as clearly separated "view model" fields layered on the contract shapes, so it stays obvious which fields have a backend source. Each gap is listed in Further Notes.
- Application status is a closed union: `running | building | failed`, with the colors from the design (green `#4EC97F`, amber `#D9A13B`, red `#E5604F`). A project's footer summary comes from its latest deploy: `deployed <rel>` or `failed <rel>`.
- Fixtures reproduce the design's sample data (masbro-labs with commerce-api, internal-tools and marketing-site, 4 members, node-01 3 of 8 GB), so the built screen can be compared visually against the design.
- Hooks return the fixture data synchronously, wrapped in the same shape a TanStack Query result would have (`data`, `isLoading`, `error`), so moving to `useQuery` later doesn't touch call sites.

**Current Organization**

- The current Organization is resolved in this order: the remembered org id from `localStorage` (key `masbro.lastOrg`, as in the design), if it is still in the user's list; otherwise the first Organization. The resolved id is written back. Every `localStorage` access is wrapped in try/catch.

**Bootstrap leftovers to remove**

These are left over from scaffolding the frontend and go away as part of this feature:

- **Placeholder dashboard:** the `Dashboard` component on the `_authed` index route that renders "Capybara Cloud / Signed in as …". The Overview replaces it. Do not keep it alongside the Overview.
- **`@supabase/supabase-js` dependency:** nothing in the frontend source imports it, and auth goes through the Go backend's `sid` cookie. Remove it from the frontend package and update the root lockfile with `pnpm install`.
- **`VITE_SUPABASE_URL` / `VITE_SUPABASE_PUBLISHABLE_KEY` in the frontend `.env`:** nothing reads them. The file is gitignored, so this is a local cleanup: remove the two keys, and delete the file if it ends up empty. This does not touch the backend's Supabase usage (local Postgres via `supabase start`).
- **Bootstrap base styles:** the `:root` rule that sets `font-family: system-ui` and `color-scheme: light dark`. The design's dark tokens and fonts replace it. The `.page`, `.form` and `.error` component styles are still used by signin/signup and stay.
- **Local build artifacts** (`dist/`, `tsconfig.tsbuildinfo`): already gitignored. Delete them locally if present; no repo change is needed.

After the cleanup, `pnpm typecheck` and `pnpm build` must still pass. The signin/signup pages must still render and submit.

**Layout and components**

- An org shell layout component holds the header and sidebar and renders the page in its main area. The Overview is the first page inside it. Later org pages (Members, Settings) will drop in as siblings.
- The components are, roughly: header (logo, breadcrumb, search, links, avatar), org sidebar (org card, nav, projects list, node footer), project card, create-project card, and status dot/chip. Initials come from a small pure helper.
- Styling uses Tailwind (already set up) with the design's palette as theme tokens: background `#131110`, surface `#1A1715`, borders `#262220`/`#2B2724`/`#3B3630`, text `#EDE8E3`/`#C9C2BA`/`#9C948A`/`#7A736B`, and the accent. Fonts: Space Grotesk and IBM Plex Mono from Google Fonts. The existing `.page` / `.form` styles for the auth pages stay.
- Layout measurements follow the design: 58px header, 248px sidebar, main content max-width 1080px, card radius 14px, grid `repeat(auto-fill, minmax(300px, 1fr))` with a 16px gap.
- The code follows the repo conventions: a docstring (`@params`, `@return`, description) on every function and component, block-form control flow only, and no multi-line conditions.

## Testing Decisions

- No automated frontend tests for this feature (user decision). The hooks return fixtures and the rest is presentational, so there is no behavior worth testing through the `createApi()` seam yet. When the hooks move to real API calls, the new `createApi()` groups (organizations, projects) get tests with the existing fake-fetch pattern in the api test file.
- No backend changes, so no backend tests.
- Verification is manual:
  - `pnpm typecheck` and `pnpm build` pass.
  - `pnpm dev` serves the Overview at `/` without signing in.
  - Searching the frontend `src` and `package.json` for `supabase` and `Signed in as` finds nothing, and `/signin` and `/signup` still render.
  - A side-by-side visual check against the design at 1440×900 covers the header, sidebar, Overview grid, hover states, all three status colors (the cron-worker and metrics-bot fixtures cover building/failed), and the empty-projects and no-organizations states (reached by editing the fixtures temporarily).

## Out of Scope

- Real authentication and session wiring, and restoring the signin redirect.
- Removing Supabase from the backend or monorepo tooling. The cleanup covers only the frontend's unused SDK and env vars.
- Any backend change: new endpoints, a member-count query, per-org project listing, application status aggregation, repo counts, or node capacity reporting.
- Project detail views: the Applications canvas with access links, and the project Members tab.
- The org Members, Git connections and Organization settings pages (rename/delete org).
- Creating Projects, the org switcher dropdown, search / ⌘K, Docs and Feedback destinations, the avatar menu.
- Mobile/responsive layouts beyond what the design's flex/grid gives by default.
- Accent color theming controls (the design's accent prop). One accent token is enough.
- Frontend automated tests.

## Further Notes

Data the design shows that the backend does not provide yet. These fields are fixture-only until a backend source exists:

- Org member count: there is no endpoint. `organization_users` has the data.
- Project environment (production/staging): not modelled anywhere.
- Per-project Applications list with status: there is no list-apps-by-project endpoint. Status would come from `application_deployments.status` / `deployment_instances.status`.
- Last deploy outcome and time per project: not exposed.
- Repo count per project: not modelled.
- Node name and memory usage: not exposed.
- `GET /api/projects` returns all of the user's projects across Organizations, not per-org, so the hook filters client-side.

There is no `CONTEXT.md` or ADR in the repo. The terms Organization, Project, Application and Deployment follow the backend's domain packages and tables.
