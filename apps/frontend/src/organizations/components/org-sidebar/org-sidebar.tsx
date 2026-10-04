import type { ReactNode } from "react";
import type { OrganizationEntry } from "../../data/types";
import { NodeStatus } from "../node-status/node-status";
import { OrgCard } from "../org-card/org-card";
import { OrgNav } from "../org-nav/org-nav";

type OrgSidebarProps = {
  entry: OrganizationEntry | null;
  projectsNav: ReactNode;
};

/**
 * @params entry: the current Organization, or null when the user has none; projectsNav: the Projects section to show under the nav
 * @return the org sidebar: org card, nav, Projects section and node footer (empty when there is no Organization)
 * The left column of every org-level page. It scrolls independently of the main area.
 */
export function OrgSidebar({ entry, projectsNav }: OrgSidebarProps) {
  if (entry === null) {
    return <aside className="w-[248px] shrink-0 border-r border-line-muted" />;
  }

  return (
    <aside className="flex w-[248px] shrink-0 flex-col gap-[22px] overflow-y-auto border-r border-line-muted px-3 py-4">
      <div className="flex flex-col gap-1">
        <OrgCard entry={entry} />
        <OrgNav memberCount={entry.view.memberCount} />
      </div>
      {projectsNav}
      <div className="flex-1" />
      <NodeStatus node={entry.view.node} />
    </aside>
  );
}
