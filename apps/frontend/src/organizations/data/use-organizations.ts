import { settledQuery, type QueryState } from "@/lib/query-state";
import { toMyOrganization } from "./contract";
import { ORGANIZATION_VIEWS, ORGANIZATIONS_RESPONSE } from "./fixtures";
import type { MyOrganization, OrganizationEntry, OrganizationView } from "./types";

const EMPTY_VIEW: OrganizationView = {
  memberCount: 0,
  node: null,
};

/**
 * @params entry: an Organization entry from the backend contract
 * @return the entry with its view-model fields attached (empty ones when the fixtures have none)
 * Layers the fixture-only view-model fields onto a contract entry.
 */
function withView(entry: MyOrganization): OrganizationEntry {
  const view = ORGANIZATION_VIEWS[entry.organization.orgId] ?? EMPTY_VIEW;
  return { ...entry, view };
}

const ORGANIZATIONS: OrganizationEntry[] = ORGANIZATIONS_RESPONSE.map(toMyOrganization).map(withView);

/**
 * @params none
 * @return the query state holding the signed-in user's Organizations
 * Lists the user's Organizations.
 * TEMPORARY: returns fixtures shaped like GET /api/organizations. Swap the body for a
 * `useQuery` over `createApi()` once auth is wired.
 */
export function useOrganizations(): QueryState<OrganizationEntry[]> {
  return settledQuery(ORGANIZATIONS);
}
