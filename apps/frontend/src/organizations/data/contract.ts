import type { MyOrganization } from "./types";

/** Mirrors dto.ListMyOrgEntry (apps/backend/pkg/dto/organization.go): one item of GET /api/organizations. */
export type ListMyOrgEntryResponse = {
  role: string;
  organization: {
    org_id: string;
    name: string;
    created_at: string;
    updated_at: string;
  };
};

/**
 * @params entry: one item of the GET /api/organizations response
 * @return the entry mapped to camelCase
 * Maps the backend's snake_case Organization entry to the frontend type.
 */
export function toMyOrganization(entry: ListMyOrgEntryResponse): MyOrganization {
  return {
    role: entry.role,
    organization: {
      orgId: entry.organization.org_id,
      name: entry.organization.name,
      createdAt: entry.organization.created_at,
      updatedAt: entry.organization.updated_at,
    },
  };
}
