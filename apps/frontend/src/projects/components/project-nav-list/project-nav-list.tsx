import { useProjects } from "../../data/use-projects";

/**
 * @params orgId: the Organization whose Projects to list
 * @return the sidebar Projects section: a heading with a "+" button and one row per Project with its app count
 * Lets the user scan every Project of the Organization from the sidebar. Rows and "+" are not functional yet.
 */
export function ProjectNavList({ orgId }: { orgId: string }) {
  const projects = useProjects(orgId);
  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center px-2.5 pb-1">
        <span className="text-[11px] font-semibold uppercase tracking-[0.08em] text-fg-subtle">Projects</span>
        <span className="flex-1" />
        <button
          type="button"
          aria-label="New project"
          className="grid size-5 cursor-pointer place-items-center rounded-[5px] text-sm text-fg-muted hover:bg-surface-strong hover:text-fg"
        >
          +
        </button>
      </div>
      {(projects.data ?? []).map((entry) => (
        <button
          key={entry.project.projectId}
          type="button"
          className="flex cursor-pointer items-center gap-[9px] rounded-[7px] px-2.5 py-[7px] text-left text-[13px] text-fg-muted hover:bg-surface-hover hover:text-fg"
        >
          <span className="w-2 text-[11px] text-fg-subtle">▸</span>
          {entry.project.name}
          <span className="flex-1" />
          <span className="font-mono text-[10.5px] text-fg-subtle">{entry.view.applications.length}</span>
        </button>
      ))}
    </div>
  );
}
