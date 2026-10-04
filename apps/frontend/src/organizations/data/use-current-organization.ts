import { useEffect } from "react";
import type { QueryState } from "@/lib/query-state";
import { readLastOrgId, writeLastOrgId } from "./last-org";
import type { OrganizationEntry } from "./types";
import { useOrganizations } from "./use-organizations";

/**
 * @params organizations: the user's Organizations; rememberedId: the remembered Organization id, if any
 * @return the remembered Organization when it is still in the list, else the first one, else null
 * Picks the Organization the dashboard shows.
 */
function resolveCurrentOrganization(organizations: OrganizationEntry[], rememberedId: string | null): OrganizationEntry | null {
  const remembered = organizations.find((entry) => entry.organization.orgId === rememberedId);
  if (remembered !== undefined) {
    return remembered;
  }
  return organizations[0] ?? null;
}

/**
 * @params none
 * @return the query state holding the current Organization (null when the user has none)
 * Resolves the current Organization from the remembered one or the first in the list,
 * and remembers the result for the next visit.
 */
export function useCurrentOrganization(): QueryState<OrganizationEntry | null> {
  const organizations = useOrganizations();
  let current: OrganizationEntry | null | undefined = undefined;
  if (organizations.data !== undefined) {
    current = resolveCurrentOrganization(organizations.data, readLastOrgId());
  }
  const currentId = current?.organization.orgId;

  useEffect(() => {
    if (currentId !== undefined) {
      writeLastOrgId(currentId);
    }
  }, [currentId]);

  return { data: current, isLoading: organizations.isLoading, error: organizations.error };
}
