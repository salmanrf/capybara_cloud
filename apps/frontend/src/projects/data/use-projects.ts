import { settledQuery, type QueryState } from "@/lib/query-state";
import { toMyProject } from "./contract";
import { PROJECT_VIEWS, PROJECTS_RESPONSE } from "./fixtures";
import type { MyProject, ProjectEntry, ProjectView } from "./types";

const EMPTY_VIEW: ProjectView = {
  environment: null,
  applications: [],
  lastDeploy: null,
  repoCount: 0,
};

/**
 * @params entry: a Project entry from the backend contract
 * @return the entry with its view-model fields attached (empty ones when the fixtures have none)
 * Layers the fixture-only view-model fields onto a contract entry.
 */
function withView(entry: MyProject): ProjectEntry {
  const view = PROJECT_VIEWS[entry.project.projectId] ?? EMPTY_VIEW;
  return { ...entry, view };
}

const PROJECTS: ProjectEntry[] = PROJECTS_RESPONSE.map(toMyProject).map(withView);

/**
 * @params orgId: the Organization whose Projects to list
 * @return the query state holding that Organization's Projects
 * Lists the Projects of one Organization. GET /api/projects returns the user's Projects
 * across every Organization, so this filters by org id on the client.
 * TEMPORARY: reads fixtures shaped like GET /api/projects. Swap the source for a
 * `useQuery` over `createApi()` once auth is wired.
 */
export function useProjects(orgId: string): QueryState<ProjectEntry[]> {
  const projects = PROJECTS.filter((entry) => entry.project.orgId === orgId);
  return settledQuery(projects);
}
