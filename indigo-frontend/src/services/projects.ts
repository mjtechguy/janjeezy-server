import { ProjectListResponseSchema, ProjectListResponse } from "@/schemas/project";

export async function fetchProjects(options?: {
  includeArchived?: boolean;
}): Promise<ProjectListResponse> {
  const params = new URLSearchParams();
  if (options?.includeArchived) {
    params.set("include_archived", "true");
  }

  const res = await fetch(
    `/api/jan/organization/projects${
      params.toString() ? `?${params.toString()}` : ""
    }`,
    {
      method: "GET",
      cache: "no-store",
    }
  );

  if (!res.ok) {
    throw new Error("Failed to load projects");
  }

  const json = await res.json();
  return ProjectListResponseSchema.parse(json);
}

export async function createProject(payload: { name: string }) {
  const res = await fetch("/api/jan/organization/projects", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Failed to create project");
  }
  return res.json();
}

export async function updateProjectName(projectPublicId: string, name: string) {
  const res = await fetch(`/api/jan/organization/projects/${projectPublicId}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Failed to update project");
  }
  return res.json();
}

export async function archiveProject(projectPublicId: string) {
  const res = await fetch(
    `/api/jan/organization/projects/${projectPublicId}/archive`,
    {
      method: "POST",
    }
  );
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Failed to archive project");
  }
  return res.json();
}
