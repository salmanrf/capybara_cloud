const LAST_ORG_KEY = "masbro.lastOrg";

/**
 * @params none
 * @return the remembered Organization id, or null when there is none or storage is unavailable
 * Reads the last Organization viewed in this browser.
 */
export function readLastOrgId(): string | null {
  try {
    return localStorage.getItem(LAST_ORG_KEY);
  } catch {
    return null;
  }
}

/**
 * @params orgId: the Organization id to remember
 * @return nothing; storage failures are ignored
 * Remembers the Organization being viewed in this browser.
 */
export function writeLastOrgId(orgId: string): void {
  try {
    localStorage.setItem(LAST_ORG_KEY, orgId);
  } catch {
    return;
  }
}
