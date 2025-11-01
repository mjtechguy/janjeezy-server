"use client";

import { useEffect, useMemo, useState } from "react";
import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  archiveProject,
  createProject,
  fetchProjects,
  updateProjectName,
} from "@/services/projects";
import { toast } from "sonner";

export default function ProjectsPage() {
  const [includeArchived, setIncludeArchived] = useState(false);
  const [projectName, setProjectName] = useState("");
  const [renameTarget, setRenameTarget] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const [renameName, setRenameName] = useState("");
  const [renameError, setRenameError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const queryKey = useMemo(
    () => ["projects", { includeArchived }],
    [includeArchived]
  );

  const { data, isLoading, error } = useQuery({
    queryKey,
    queryFn: () => fetchProjects({ includeArchived }),
  });

  const createMutation = useMutation({
    mutationFn: createProject,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey });
      toast.success("Project created");
      setProjectName("");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const renameMutation = useMutation({
    mutationFn: ({ publicId, name }: { publicId: string; name: string }) =>
      updateProjectName(publicId, name),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey });
      toast.success("Project updated");
      setRenameTarget(null);
      setRenameName("");
      setRenameError(null);
    },
    onError: (err: Error) => {
      toast.error(err.message);
      setRenameError(err.message);
    },
  });

  const archiveMutation = useMutation({
    mutationFn: (publicId: string) => archiveProject(publicId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey });
      toast.success("Project archived");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const handleCreateProject = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!projectName.trim()) {
      toast.error("Project name is required");
      return;
    }
    createMutation.mutate({ name: projectName.trim() });
  };

  const closeRenameModal = () => {
    if (renameMutation.isPending) {
      return;
    }
    setRenameTarget(null);
    setRenameName("");
    setRenameError(null);
  };

  const handleRenameSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!renameTarget) {
      return;
    }
    const nextName = renameName.trim();
    if (!nextName) {
      setRenameError("Project name is required");
      return;
    }
    if (nextName === renameTarget.name) {
      setRenameError("Enter a different name to update");
      return;
    }
    setRenameError(null);
    renameMutation.mutate({ publicId: renameTarget.id, name: nextName });
  };

  useEffect(() => {
    if (!renameTarget) {
      return;
    }
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !renameMutation.isPending) {
        setRenameTarget(null);
        setRenameName("");
        setRenameError(null);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [renameTarget, renameMutation.isPending]);

  return (
    <section className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-12">
      <header className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-foreground">Projects</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Browse organization projects. In future iterations this surface will
            enable creation, updates, and provider assignments.
          </p>
        </div>

        <label className="inline-flex items-center gap-2 text-sm text-muted-foreground">
          <input
            type="checkbox"
            className="h-4 w-4 rounded border-border text-primary focus:ring-[#0EA5E9]"
            checked={includeArchived}
            onChange={(event) => {
              setIncludeArchived(event.target.checked);
            }}
          />
          Include archived
        </label>
      </header>

      <form
        onSubmit={handleCreateProject}
        className="flex flex-col gap-3 rounded-2xl border border-border bg-card p-6 shadow-sm sm:flex-row sm:items-end"
      >
        <div className="flex-1">
          <label className="text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
            Project name
          </label>
          <input
            type="text"
            value={projectName}
            onChange={(event) => setProjectName(event.target.value)}
            placeholder="Internal automation"
            className="mt-2 w-full rounded-xl border border-border px-4 py-2 text-sm focus-visible:ring-ring focus-visible:border-transparent focus:outline-none"
            disabled={createMutation.isPending}
          />
        </div>
        <button
          type="submit"
          className="h-10 rounded-xl bg-primary px-4 text-sm font-semibold text-primary-foreground transition hover:brightness-110 disabled:opacity-50"
          disabled={createMutation.isPending}
        >
          {createMutation.isPending ? "Creating…" : "Create project"}
        </button>
      </form>

      <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
        <table className="min-w-full divide-y divide-border/60">
          <thead className="bg-muted text-left text-xs font-semibold uppercase tracking-[0.35em] text-muted-foreground">
            <tr>
              <th className="px-6 py-3">Name</th>
              <th className="px-6 py-3">Status</th>
              <th className="px-6 py-3">Created</th>
              <th className="px-6 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40 text-sm">
            {isLoading && (
              <tr>
                <td className="px-6 py-4 text-muted-foreground" colSpan={3}>
                  Loading projects…
                </td>
              </tr>
            )}
            {error && !isLoading && (
              <tr>
                <td className="px-6 py-4 text-danger" colSpan={3}>
                  Failed to load projects. Please try again later.
                </td>
              </tr>
            )}
            {data?.data.map((project) => (
              <tr key={project.id} className="hover:bg-muted/60">
                <td className="px-6 py-4 font-medium text-foreground">
                  {project.name}
                </td>
                <td className="px-6 py-4">
                  <span
                    className={`inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold ${
                      project.status === "active"
                        ? "bg-success/10 text-success"
                        : "bg-warning/10 text-warning"
                    }`}
                  >
                    {project.status}
                  </span>
                </td>
                <td className="px-6 py-4 text-muted-foreground">
                  {new Date(project.created_at * 1000).toLocaleDateString()}
                </td>
                <td className="px-6 py-4 text-right">
                  <div className="flex justify-end gap-2">
                    <button
                      type="button"
                      className="rounded-lg border border-border px-3 py-1 text-xs font-semibold text-muted-foreground transition hover:border-border hover:text-foreground"
                      onClick={() => {
                        setRenameTarget({ id: project.id, name: project.name });
                        setRenameName(project.name);
                        setRenameError(null);
                      }}
                      disabled={renameMutation.isPending}
                    >
                      Rename
                    </button>
                    <button
                      type="button"
                      className="rounded-lg border border-transparent px-3 py-1 text-xs font-semibold text-danger transition hover:bg-danger/10 disabled:opacity-50"
                      onClick={() => archiveMutation.mutate(project.id)}
                      disabled={
                        archiveMutation.isPending ||
                        project.status === "archived"
                      }
                    >
                      Archive
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {!isLoading && !error && data?.data.length === 0 && (
              <tr>
                <td className="px-6 py-4 text-muted-foreground" colSpan={4}>
                  No projects yet. Use the form above to create your first
                  project.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      {renameTarget && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-background/80 px-4"
          onClick={closeRenameModal}
        >
          <div
            className="w-full max-w-md rounded-2xl border border-border bg-card p-6 shadow-xl"
            onClick={(event) => event.stopPropagation()}
          >
            <h2 className="text-lg font-semibold text-foreground">
              Rename project
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              Update the display name for this project. The project ID remains
              unchanged.
            </p>
            <form className="mt-4 flex flex-col gap-4" onSubmit={handleRenameSubmit}>
              <div>
                <label className="text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
                  New name
                </label>
                <input
                  type="text"
                  value={renameName}
                  onChange={(event) => {
                    setRenameName(event.target.value);
                    setRenameError(null);
                  }}
                  className={`mt-2 w-full rounded-xl border px-4 py-2 text-sm focus:outline-none ${
                    renameError
                      ? "border-danger focus:border-danger"
                      : "border-border focus-visible:ring-ring focus-visible:border-transparent"
                  }`}
                  disabled={renameMutation.isPending}
                  autoFocus
                />
                {renameError && (
                  <p className="mt-2 text-xs text-danger">{renameError}</p>
                )}
              </div>
              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  className="rounded-lg border border-border px-3 py-2 text-sm font-semibold text-muted-foreground transition hover:border-border hover:text-foreground disabled:opacity-50"
                  onClick={closeRenameModal}
                  disabled={renameMutation.isPending}
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground transition hover:brightness-110 disabled:opacity-50"
                  disabled={renameMutation.isPending}
                >
                  {renameMutation.isPending ? "Saving…" : "Save changes"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </section>
  );
}
