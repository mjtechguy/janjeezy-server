"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchOrganizationMembers,
  updateOrganizationMemberRole,
} from "@/services/organization";
import { toast } from "sonner";

export default function UsersPage() {
  const queryClient = useQueryClient();
  const membersQuery = useQuery({
    queryKey: ["organization-members"],
    queryFn: fetchOrganizationMembers,
  });

  const updateRoleMutation = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: "owner" | "reader" }) =>
      updateOrganizationMemberRole(userId, role),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["organization-members"] });
      toast.success("Member role updated");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  return (
    <section className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-12">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold text-foreground">
          Organization Members
        </h1>
        <p className="text-sm text-muted-foreground">
          Manage member access and elevate readers to owners when needed. Role
          changes take effect immediately.
        </p>
      </header>

      <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
        <table className="min-w-full divide-y divide-border/60">
          <thead className="bg-muted text-left text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
            <tr>
              <th className="px-6 py-3">Name</th>
              <th className="px-6 py-3">Email</th>
              <th className="px-6 py-3">Joined</th>
              <th className="px-6 py-3">Role</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40 text-sm text-muted-foreground">
            {membersQuery.isLoading && (
              <tr>
                <td className="px-6 py-4" colSpan={4}>
                  Loading members…
                </td>
              </tr>
            )}
            {membersQuery.error && !membersQuery.isLoading && (
              <tr>
                <td className="px-6 py-4 text-danger" colSpan={4}>
                  Failed to load organization members.
                </td>
              </tr>
            )}
            {membersQuery.data?.data.map((member) => (
              <tr key={member.user.id} className="hover:bg-muted/60">
                <td className="px-6 py-4 font-medium text-foreground">
                  {member.user.name}
                </td>
                <td className="px-6 py-4">
                  {member.user.email}
                </td>
                <td className="px-6 py-4">
                  {new Date(member.user.created_at * 1000).toLocaleDateString()}
                </td>
                <td className="px-6 py-4">
                  <select
                    value={member.role}
                    onChange={(event) =>
                      updateRoleMutation.mutate({
                        userId: member.user.id,
                        role: event.target.value as "owner" | "reader",
                      })
                    }
                    className="rounded-xl border border-border bg-background px-3 py-1 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    disabled={updateRoleMutation.isPending}
                  >
                    <option value="owner">Owner</option>
                    <option value="reader">Reader</option>
                  </select>
                </td>
              </tr>
            ))}
            {!membersQuery.isLoading &&
              !membersQuery.error &&
              (membersQuery.data?.data.length ?? 0) === 0 && (
                <tr>
                  <td className="px-6 py-4" colSpan={4}>
                    No members found.
                  </td>
                </tr>
              )}
          </tbody>
        </table>
      </div>
    </section>
  );
}
