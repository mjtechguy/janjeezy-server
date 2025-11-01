"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createAdminApiKey,
  deleteAdminApiKey,
  fetchAdminApiKeys,
} from "@/services/api-keys";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function ApiKeysPage() {
  const queryClient = useQueryClient();
  const apiKeyQuery = useQuery({
    queryKey: ["admin-api-keys"],
    queryFn: fetchAdminApiKeys,
  });

  const [keyName, setKeyName] = useState("");
  const [generatedKey, setGeneratedKey] = useState<string | null>(null);

  const createMutation = useMutation({
    mutationFn: createAdminApiKey,
    onSuccess: async (data) => {
      await queryClient.invalidateQueries({ queryKey: ["admin-api-keys"] });
      setGeneratedKey(data.value ?? null);
      toast.success("API key created");
      setKeyName("");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const deleteMutation = useMutation({
    mutationFn: deleteAdminApiKey,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["admin-api-keys"] });
      toast.success("API key deleted");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  return (
    <section className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-12">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold text-foreground">
          Admin API Keys
        </h1>
        <p className="text-sm text-muted-foreground">
          Generate administrative API keys for automation or integrate with
          trusted systems. Keys are shown only once after creation.
        </p>
      </header>

      <form
        onSubmit={(event) => {
          event.preventDefault();
          if (!keyName.trim()) {
            toast.error("Name is required");
            return;
          }
          createMutation.mutate(keyName.trim());
        }}
        className="flex flex-col gap-3 rounded-2xl border border-border bg-card p-6 shadow-sm sm:flex-row sm:items-end"
      >
        <div className="flex-1">
          <Label className="text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
            Key name
          </Label>
          <Input
            type="text"
            value={keyName}
            onChange={(event) => setKeyName(event.target.value)}
            placeholder="CI deployment"
            className="mt-2"
            disabled={createMutation.isPending}
          />
        </div>
        <Button type="submit" disabled={createMutation.isPending}>
          {createMutation.isPending ? "Creating…" : "Create API key"}
        </Button>
      </form>

      {generatedKey && (
        <div className="rounded-2xl border border-border bg-muted/40 p-6 text-sm text-muted-foreground">
          <p className="font-semibold">New API key</p>
          <p className="mt-2 break-all font-mono text-xs">{generatedKey}</p>
          <p className="mt-2 text-xs text-muted-foreground">
            Copy and store this value securely. You will not be able to see it
            again.
          </p>
        </div>
      )}

      <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
        <table className="min-w-full divide-y divide-border/60">
          <thead className="bg-muted text-left text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
            <tr>
              <th className="px-6 py-3">Name</th>
              <th className="px-6 py-3">Created</th>
              <th className="px-6 py-3">Last used</th>
              <th className="px-6 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40 text-sm text-muted-foreground">
            {apiKeyQuery.isLoading && (
              <tr>
                <td className="px-6 py-4" colSpan={4}>
                  Loading API keys…
                </td>
              </tr>
            )}
            {apiKeyQuery.error && !apiKeyQuery.isLoading && (
              <tr>
                <td className="px-6 py-4 text-danger" colSpan={4}>
                  Failed to load API keys.
                </td>
              </tr>
            )}
            {apiKeyQuery.data?.data.map((key) => (
              <tr key={key.id} className="hover:bg-muted/60">
                <td className="px-6 py-4 font-medium text-foreground">
                  {key.name ?? key.id}
                </td>
                <td className="px-6 py-4">
                  {new Date(key.created_at * 1000).toLocaleDateString()}
                </td>
                <td className="px-6 py-4">
                  {key.last_used_at
                    ? new Date(key.last_used_at * 1000).toLocaleDateString()
                    : "—"}
                </td>
                <td className="px-6 py-4 text-right">
                  <button
                    type="button"
                    className="rounded-lg border border-danger/40 px-3 py-1 text-xs font-semibold text-danger transition hover:bg-danger/10 disabled:opacity-50"
                    onClick={() => deleteMutation.mutate(key.id)}
                    disabled={deleteMutation.isPending}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
            {!apiKeyQuery.isLoading &&
              !apiKeyQuery.error &&
              (apiKeyQuery.data?.data.length ?? 0) === 0 && (
                <tr>
                  <td className="px-6 py-4" colSpan={4}>
                    No admin API keys found.
                  </td>
                </tr>
              )}
          </tbody>
        </table>
      </div>
    </section>
  );
}
