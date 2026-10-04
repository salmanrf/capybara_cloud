import type { MyProject } from "./types";

/** Mirrors dto.ListMyProjectEntry (apps/backend/pkg/dto/project.go): one item of GET /api/projects. */
export type ListMyProjectEntryResponse = {
  role: string;
  project: {
    org_id: string;
    project_id: string;
    name: string;
    created_at: string;
    updated_at: string;
  };
};

/**
 * @params entry: one item of the GET /api/projects response
 * @return the entry mapped to camelCase
 * Maps the backend's snake_case Project entry to the frontend type.
 */
export function toMyProject(entry: ListMyProjectEntryResponse): MyProject {
  return {
    role: entry.role,
    project: {
      orgId: entry.project.org_id,
      projectId: entry.project.project_id,
      name: entry.project.name,
      createdAt: entry.project.created_at,
      updatedAt: entry.project.updated_at,
    },
  };
}
