/** An Organization as the backend returns it, in camelCase. */
export type Organization = {
  orgId: string;
  name: string;
  createdAt: string;
  updatedAt: string;
};

/** One entry of the signed-in user's Organizations: their role and the Organization. */
export type MyOrganization = {
  role: string;
  organization: Organization;
};

/** A host node's memory usage. */
export type NodeUsage = {
  name: string;
  usedGb: number;
  totalGb: number;
};

/**
 * View-model fields the design shows that no backend endpoint provides yet.
 * They are fixture-only until a backend source exists.
 */
export type OrganizationView = {
  memberCount: number;
  node: NodeUsage | null;
};

/** An Organization entry as the UI reads it: the backend contract plus the view-model fields. */
export type OrganizationEntry = MyOrganization & {
  view: OrganizationView;
};
