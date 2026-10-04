import type { ReactNode } from "react";
import type { User } from "@/api";
import { NoOrganizations } from "../components/no-organizations/no-organizations";
import { OrgHeader } from "../components/org-header/org-header";
import { OrgSidebar } from "../components/org-sidebar/org-sidebar";
import { useCurrentOrganization } from "../data/use-current-organization";
import type { OrganizationEntry } from "../data/types";

type OrgShellProps = {
  user: User;
  renderProjectsNav: (orgId: string) => ReactNode;
  children: (entry: OrganizationEntry) => ReactNode;
};

/**
 * @params user: the signed-in user; renderProjectsNav: builds the sidebar Projects section for an org id;
 *   children: builds the page for the current Organization
 * @return the full-height org layout: fixed header, sidebar and a scrolling main area
 * The shell of every org-level page. It resolves the current Organization and shows an
 * empty state instead of the page when the user has none.
 */
export function OrgShell({ user, renderProjectsNav, children }: OrgShellProps) {
  const current = useCurrentOrganization();
  const entry = current.data ?? null;

  let projectsNav: ReactNode = null;
  let page: ReactNode = <NoOrganizations />;
  if (entry !== null) {
    projectsNav = renderProjectsNav(entry.organization.orgId);
    page = children(entry);
  }

  return (
    <div className="flex h-screen flex-col overflow-hidden">
      <OrgHeader user={user} organization={entry?.organization ?? null} />
      <div className="flex min-h-0 flex-1">
        <OrgSidebar entry={entry} projectsNav={projectsNav} />
        <main className="min-w-0 flex-1 overflow-auto">
          <div className="mx-auto flex max-w-[1080px] flex-col gap-6 px-9 pt-8 pb-12">{page}</div>
        </main>
      </div>
    </div>
  );
}
