import type { Organization } from "../../data/types";

/**
 * @params organization: the Organization the Overview is for
 * @return the page heading: an "Overview" eyebrow over the org name, and the "New Project" button
 * Labels the Overview page. "New Project" is not functional yet.
 */
export function OrgOverviewHeader({ organization }: { organization: Organization }) {
  return (
    <div className="flex items-end gap-3.5">
      <div className="flex flex-col gap-1.5">
        <span className="text-xs font-semibold uppercase tracking-[0.08em] text-fg-subtle">Overview</span>
        <h1 className="m-0 text-2xl font-semibold tracking-[-0.02em]">{organization.name}</h1>
      </div>
      <div className="flex-1" />
      <button
        type="button"
        className="shrink-0 cursor-pointer whitespace-nowrap rounded-lg bg-accent px-4 py-2 text-[13px] font-semibold text-on-accent hover:brightness-110"
      >
        New Project
      </button>
    </div>
  );
}
