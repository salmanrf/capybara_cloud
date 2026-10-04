import { formatRelativeTime, pluralize } from "@/lib/format";
import type { LastDeploy, ProjectEntry } from "../../data/types";
import { ApplicationChip } from "../application-chip/application-chip";
import { EnvironmentBadge } from "../environment-badge/environment-badge";

/**
 * @params lastDeploy: the Project's latest deploy, or null when it was never deployed
 * @return the footer summary: "deployed 12m ago", "failed 1h ago" or "no deploys yet"
 * Describes a Project's latest deploy.
 */
function describeLastDeploy(lastDeploy: LastDeploy | null): string {
  if (lastDeploy === null) {
    return "no deploys yet";
  }
  return `${lastDeploy.outcome} ${formatRelativeTime(lastDeploy.at)}`;
}

/**
 * @params entry: the Project to show
 * @return a card with the Project's name, environment, Applications and deploy summary
 * Summarises one Project on the Overview grid.
 */
export function ProjectCard({ entry }: { entry: ProjectEntry }) {
  const { project, view } = entry;
  return (
    <article className="flex cursor-pointer flex-col gap-4 rounded-[14px] border border-line bg-surface p-5 hover:border-line-strong hover:bg-surface-hover">
      <div className="flex items-center gap-2.5">
        <span className="text-[15px] font-semibold">{project.name}</span>
        <div className="flex-1" />
        <EnvironmentBadge environment={view.environment} />
      </div>
      <div className="flex flex-wrap gap-1.5">
        {view.applications.map((application) => (
          <ApplicationChip key={application.name} application={application} />
        ))}
      </div>
      <div className="flex items-center gap-2 border-t border-surface-strong pt-3 text-xs text-fg-subtle">
        <span>{pluralize(view.applications.length, "app")}</span>
        <span>·</span>
        <span>{describeLastDeploy(view.lastDeploy)}</span>
        <span className="flex-1" />
        <span className="whitespace-nowrap font-mono text-[11px]">{pluralize(view.repoCount, "repo")}</span>
      </div>
    </article>
  );
}
