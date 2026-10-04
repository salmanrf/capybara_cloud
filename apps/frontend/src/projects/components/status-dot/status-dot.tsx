import type { ApplicationStatus } from "../../data/types";

const STATUS_COLOR: Record<ApplicationStatus, string> = {
  running: "bg-status-running",
  building: "bg-status-building",
  failed: "bg-status-failed",
};

/**
 * @params status: the Application's status
 * @return a small dot in the status color (green running, amber building, red failed)
 * Shows an Application's status at a glance.
 */
export function StatusDot({ status }: { status: ApplicationStatus }) {
  return <span role="img" aria-label={status} className={`size-1.5 shrink-0 rounded-full ${STATUS_COLOR[status]}`} />;
}
