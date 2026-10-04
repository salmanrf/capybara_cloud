import type { ProjectApplication } from "../../data/types";
import { StatusDot } from "../status-dot/status-dot";

/**
 * @params application: the Application to show
 * @return a chip with the Application's status dot and name
 * Lists one Application on a Project card.
 */
export function ApplicationChip({ application }: { application: ProjectApplication }) {
  return (
    <span className="flex items-center gap-1.5 whitespace-nowrap rounded-md bg-surface-raised px-2 py-1 font-mono text-[11px] text-fg-secondary">
      <StatusDot status={application.status} />
      {application.name}
    </span>
  );
}
