import { createFileRoute } from "@tanstack/react-router";
import { useAuth } from "@/auth/data/use-auth";
import { OrgOverviewHeader } from "@/organizations/components/org-overview-header/org-overview-header";
import { OrgShell } from "@/organizations/layout/org-shell";
import { ProjectGrid } from "@/projects/components/project-grid/project-grid";
import { ProjectNavList } from "@/projects/components/project-nav-list/project-nav-list";

export const Route = createFileRoute("/_authed/")({
  component: Overview,
});

/**
 * @params none
 * @return the org Overview page inside the org shell
 * Landing page: the current Organization's name and its Projects.
 */
function Overview() {
  const { user } = useAuth();
  return (
    <OrgShell user={user} renderProjectsNav={(orgId) => <ProjectNavList orgId={orgId} />}>
      {(entry) => (
        <>
          <OrgOverviewHeader organization={entry.organization} />
          <ProjectGrid orgId={entry.organization.orgId} />
        </>
      )}
    </OrgShell>
  );
}
