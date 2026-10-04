import type { ListMyProjectEntryResponse } from "./contract";
import type { ProjectView } from "./types";

const MASBRO_LABS_ID = "3f6a1c9e-2b7d-4e58-9a0c-6d4b8e1f2a53";
const COMMERCE_API_ID = "a2c4e6f8-1b3d-4f5a-8c7e-9d0b2a4c6e81";
const INTERNAL_TOOLS_ID = "b3d5f7a9-2c4e-4a6b-9d8f-0e1c3b5d7f92";
const MARKETING_SITE_ID = "c4e6a8b0-3d5f-4b7c-8e9a-1f2d4c6e8a03";

const MINUTE_MS = 60_000;

/**
 * @params minutes: how many minutes ago
 * @return the Date that many minutes before now
 * Anchors fixture times to page load, so relative times match the design.
 */
function minutesAgo(minutes: number): Date {
  return new Date(Date.now() - minutes * MINUTE_MS);
}

/** TEMPORARY: stands in for the GET /api/projects response until the hook calls the backend. */
export const PROJECTS_RESPONSE: ListMyProjectEntryResponse[] = [
  {
    role: "owner",
    project: {
      org_id: MASBRO_LABS_ID,
      project_id: COMMERCE_API_ID,
      name: "commerce-api",
      created_at: "2026-06-03T10:00:00Z",
      updated_at: "2026-06-03T10:00:00Z",
    },
  },
  {
    role: "owner",
    project: {
      org_id: MASBRO_LABS_ID,
      project_id: INTERNAL_TOOLS_ID,
      name: "internal-tools",
      created_at: "2026-06-10T14:30:00Z",
      updated_at: "2026-06-10T14:30:00Z",
    },
  },
  {
    role: "owner",
    project: {
      org_id: MASBRO_LABS_ID,
      project_id: MARKETING_SITE_ID,
      name: "marketing-site",
      created_at: "2026-07-01T08:45:00Z",
      updated_at: "2026-07-01T08:45:00Z",
    },
  },
];

/** Fixture-only view-model fields, keyed by project_id. No backend endpoint provides them yet. */
export const PROJECT_VIEWS: Record<string, ProjectView> = {
  [COMMERCE_API_ID]: {
    environment: "production",
    applications: [
      { name: "api-gateway", status: "running" },
      { name: "storefront-web", status: "running" },
      { name: "cron-worker", status: "building" },
    ],
    lastDeploy: { outcome: "deployed", at: minutesAgo(12) },
    repoCount: 3,
  },
  [INTERNAL_TOOLS_ID]: {
    environment: "production",
    applications: [
      { name: "admin-panel", status: "running" },
      { name: "metrics-bot", status: "failed" },
    ],
    lastDeploy: { outcome: "failed", at: minutesAgo(60) },
    repoCount: 2,
  },
  [MARKETING_SITE_ID]: {
    environment: "staging",
    applications: [{ name: "landing", status: "running" }],
    lastDeploy: { outcome: "deployed", at: minutesAgo(3 * 24 * 60) },
    repoCount: 1,
  },
};
