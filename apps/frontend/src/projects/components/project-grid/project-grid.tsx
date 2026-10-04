import { useProjects } from "../../data/use-projects";
import { CreateProjectCard } from "../create-project-card/create-project-card";
import { ProjectCard } from "../project-card/project-card";

/**
 * @params orgId: the Organization whose Projects to show
 * @return a responsive grid of Project cards ending with the "Create a project" card
 * Lays out an Organization's Projects on the Overview. With no Projects only the create card shows.
 */
export function ProjectGrid({ orgId }: { orgId: string }) {
  const projects = useProjects(orgId);
  return (
    <div className="grid grid-cols-[repeat(auto-fill,minmax(300px,1fr))] gap-4">
      {(projects.data ?? []).map((entry) => (
        <ProjectCard key={entry.project.projectId} entry={entry} />
      ))}
      <CreateProjectCard />
    </div>
  );
}
