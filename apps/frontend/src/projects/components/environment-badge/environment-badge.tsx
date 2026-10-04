/**
 * @params environment: the Project's environment, or null when it is unknown
 * @return a pill with the environment name, or null when it is unknown
 * Tells Projects apart by environment (e.g. production, staging).
 */
export function EnvironmentBadge({ environment }: { environment: string | null }) {
  if (environment === null) {
    return null;
  }
  return <span className="rounded-full border border-line px-[9px] py-0.5 text-[11px] text-fg-muted">{environment}</span>;
}
