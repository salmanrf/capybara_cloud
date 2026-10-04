import type { ListMyOrgEntryResponse } from "./contract";
import type { OrganizationView } from "./types";

const MASBRO_LABS_ID = "3f6a1c9e-2b7d-4e58-9a0c-6d4b8e1f2a53";

/** TEMPORARY: stands in for the GET /api/organizations response until the hook calls the backend. */
export const ORGANIZATIONS_RESPONSE: ListMyOrgEntryResponse[] = [
  {
    role: "owner",
    organization: {
      org_id: MASBRO_LABS_ID,
      name: "masbro-labs",
      created_at: "2026-06-02T09:14:00Z",
      updated_at: "2026-06-02T09:14:00Z",
    },
  },
];

/** Fixture-only view-model fields, keyed by org_id. No backend endpoint provides them yet. */
export const ORGANIZATION_VIEWS: Record<string, OrganizationView> = {
  [MASBRO_LABS_ID]: {
    memberCount: 4,
    node: { name: "node-01", usedGb: 3, totalGb: 8 },
  },
};
