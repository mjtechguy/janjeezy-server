"use client";

import { useMemo, useState } from "react";
import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";
import { fetchInvites, createInvite, deleteInvite } from "@/services/invites";
import { fetchProjects } from "@/services/projects";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";

type ProjectRole = "member" | "owner";
type InviteRole = "reader" | "owner";

export default function InvitesPage() {
  const queryClient = useQueryClient();
  const invitesQuery = useQuery({
    queryKey: ["organization-invites"],
    queryFn: fetchInvites,
  });
  const projectsQuery = useQuery({
    queryKey: ["projects", { includeArchived: false }],
    queryFn: () => fetchProjects({ includeArchived: false }),
  });

  const [email, setEmail] = useState("");
  const [role, setRole] = useState<InviteRole>("reader");
  const [selectedProjects, setSelectedProjects] = useState<
    Record<string, ProjectRole>
  >({});

  const invites = invitesQuery.data?.data ?? [];
  const projects = useMemo(
    () => projectsQuery.data?.data ?? [],
    [projectsQuery.data]
  );

  const createMutation = useMutation({
    mutationFn: createInvite,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["organization-invites"] });
      toast.success("Invite created");
      setEmail("");
      setRole("reader");
      setSelectedProjects({});
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteInvite,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["organization-invites"] });
      toast.success("Invite deleted");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const toggleProject = (projectId: string) => {
    setSelectedProjects((prev) => {
      const copy = { ...prev };
      if (projectId in copy) {
        delete copy[projectId];
      } else {
        copy[projectId] = "member";
      }
      return copy;
    });
  };

  const changeProjectRole = (projectId: string, nextRole: ProjectRole) => {
    setSelectedProjects((prev) => ({
      ...prev,
      [projectId]: nextRole,
    }));
  };

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const trimmedEmail = email.trim();
    if (!trimmedEmail) {
      toast.error("Email is required");
      return;
    }

    const projectsPayload = Object.entries(selectedProjects).map(
      ([id, projectRole]) => ({
        id,
        role: projectRole,
      })
    );

    createMutation.mutate({
      email: trimmedEmail,
      role,
      projects: projectsPayload.length > 0 ? projectsPayload : [],
    });
  };

  const isCreating = createMutation.isPending;

  return (
    <section className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-12">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold text-foreground">Invites</h1>
        <p className="text-sm text-muted-foreground">
          Send new invitations, track pending requests, and revoke unused codes
          for your organization.
        </p>
      </header>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-semibold">
            Create invitation
          </CardTitle>
          <CardDescription>
            Recipients receive an email with a one-time link generated from your
            SMTP configuration.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="grid gap-6 lg:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="invite-email">Email</Label>
              <Input
                id="invite-email"
                type="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                placeholder="teammate@example.com"
                required
                disabled={isCreating}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="invite-role">Organization role</Label>
              <select
                id="invite-role"
                value={role}
                onChange={(event) => setRole(event.target.value as InviteRole)}
                className="h-10 rounded-md border border-input bg-background px-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                disabled={isCreating}
              >
                <option value="reader">Reader</option>
                <option value="owner">Owner</option>
              </select>
            </div>

            <div className="lg:col-span-2 space-y-4">
              <div className="space-y-2">
                <Label>Project access (optional)</Label>
                <p className="text-xs text-muted-foreground">
                  Select projects to pre-provision membership. Readers receive
                  member access by default.
                </p>
              </div>

              <div className="grid gap-3 sm:grid-cols-2">
                {projects.length === 0 && (
                  <p className="text-sm text-muted-foreground">
                    No projects available yet. Create a project first to assign it
                    to an invite.
                  </p>
                )}
                {projects.map((project) => {
                  const checked = project.id in selectedProjects;
                  const currentRole = selectedProjects[project.id] ?? "member";
                  return (
                    <div
                      key={project.id}
                      className="rounded-lg border border-border bg-card/80 p-4"
                    >
                      <label className="flex items-center gap-3 text-sm font-medium text-foreground">
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={() => toggleProject(project.id)}
                          className="h-4 w-4 rounded border-input text-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                          disabled={isCreating}
                        />
                        {project.name}
                      </label>
                      <p className="mt-1 text-xs text-muted-foreground">
                        {project.id}
                      </p>
                      {checked && (
                        <div className="mt-3 space-y-1">
                          <Label
                            htmlFor={`${project.id}-role`}
                            className="text-xs uppercase tracking-[0.2em] text-muted-foreground"
                          >
                            Project role
                          </Label>
                          <select
                            id={`${project.id}-role`}
                            value={currentRole}
                            onChange={(event) =>
                              changeProjectRole(
                                project.id,
                                event.target.value as ProjectRole
                              )
                            }
                            className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                            disabled={isCreating}
                          >
                            <option value="member">Member</option>
                            <option value="owner">Owner</option>
                          </select>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>

              <div className="flex justify-end">
                <Button type="submit" disabled={isCreating}>
                  {isCreating ? "Sending…" : "Send invite"}
                </Button>
              </div>
            </div>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-semibold">
            Pending & recent invites
          </CardTitle>
          <CardDescription>
            Track the current status of invite codes and revoke them when no
            longer needed.
          </CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-border/60 text-sm">
              <thead className="bg-muted text-left text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
                <tr>
                  <th className="px-6 py-3">Email</th>
                  <th className="px-6 py-3">Role</th>
                  <th className="px-6 py-3">Status</th>
                  <th className="px-6 py-3">Projects</th>
                  <th className="px-6 py-3">Invited</th>
                  <th className="px-6 py-3">Expires</th>
                  <th className="px-6 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/40 text-muted-foreground">
                {invitesQuery.isLoading && (
                  <tr>
                    <td className="px-6 py-4" colSpan={7}>
                      Loading invites…
                    </td>
                  </tr>
                )}
                {invitesQuery.error && !invitesQuery.isLoading && (
                  <tr>
                    <td className="px-6 py-4 text-danger" colSpan={7}>
                      Failed to load invites.
                    </td>
                  </tr>
                )}
                {invites.map((invite) => {
                  const projectsLabel =
                    invite.projects.length === 0
                      ? "—"
                      : invite.projects
                          .map((proj) => `${proj.id} (${proj.role})`)
                          .join(", ");
                  const isPending = invite.status === "pending";
                  return (
                    <tr key={invite.id} className="hover:bg-muted/50">
                      <td className="px-6 py-4 font-medium text-foreground">
                        {invite.email}
                      </td>
                      <td className="px-6 py-4">{invite.role}</td>
                      <td className="px-6 py-4">
                        <Badge
                          variant={
                            invite.status === "pending"
                              ? "secondary"
                              : invite.status === "accepted"
                              ? "default"
                              : "destructive"
                          }
                        >
                          {invite.status}
                        </Badge>
                      </td>
                      <td className="px-6 py-4">{projectsLabel}</td>
                      <td className="px-6 py-4">
                        {new Date(invite.invited_at).toLocaleString()}
                      </td>
                      <td className="px-6 py-4">
                        {new Date(invite.expires_at).toLocaleString()}
                      </td>
                      <td className="px-6 py-4 text-right">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => deleteMutation.mutate(invite.id)}
                          disabled={!isPending || deleteMutation.isPending}
                        >
                          Revoke
                        </Button>
                      </td>
                    </tr>
                  );
                })}
                {!invitesQuery.isLoading &&
                  !invitesQuery.error &&
                  invites.length === 0 && (
                    <tr>
                      <td className="px-6 py-4" colSpan={7}>
                        No invites have been sent yet.
                      </td>
                    </tr>
                  )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </section>
  );
}
