/** A Project as the backend returns it, in camelCase. */
export type Project = {
  orgId: string;
  projectId: string;
  name: string;
  createdAt: string;
  updatedAt: string;
};

/** One entry of the signed-in user's Projects: their role and the Project. */
export type MyProject = {
  role: string;
  project: Project;
};

/** The state of an Application's latest deployment. */
export type ApplicationStatus = "running" | "building" | "failed";

/** An Application inside a Project, as the Project card lists it. */
export type ProjectApplication = {
  name: string;
  status: ApplicationStatus;
};

/** The outcome and time of a Project's latest deploy. */
export type LastDeploy = {
  outcome: "deployed" | "failed";
  at: Date;
};

/**
 * View-model fields the design shows that no backend endpoint provides yet.
 * They are fixture-only until a backend source exists.
 */
export type ProjectView = {
  environment: string | null;
  applications: ProjectApplication[];
  lastDeploy: LastDeploy | null;
  repoCount: number;
};

/** A Project entry as the UI reads it: the backend contract plus the view-model fields. */
export type ProjectEntry = MyProject & {
  view: ProjectView;
};
