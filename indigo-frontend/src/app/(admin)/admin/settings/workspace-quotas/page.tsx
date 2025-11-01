"use client";

import { useEffect, useState, useTransition } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import {
  fetchWorkspaceQuota,
  updateWorkspaceQuota,
} from "@/services/settings";
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

type OverrideState = {
  user_public_id: string;
  limit: number;
};

export default function WorkspaceQuotaPage() {
  const queryClient = useQueryClient();
  const [defaultLimit, setDefaultLimit] = useState(10);
  const [overrides, setOverrides] = useState<OverrideState[]>([]);
  const [_isTransitionPending, startTransition] = useTransition();

  const quotaQuery = useQuery({
    queryKey: ["workspace-quotas"],
    queryFn: fetchWorkspaceQuota,
  });

  useEffect(() => {
    if (!quotaQuery.data) {
      return;
    }
    startTransition(() => {
      setDefaultLimit(quotaQuery.data.default_limit);
      setOverrides(quotaQuery.data.overrides);
    });
  }, [quotaQuery.data, startTransition]);

  const mutation = useMutation({
    mutationFn: updateWorkspaceQuota,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["workspace-quotas"] });
      toast.success("Workspace quotas updated");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (defaultLimit <= 0) {
      toast.error("Default limit must be a positive number");
      return;
    }
    const sanitized = overrides.filter((entry) => entry.user_public_id.trim() !== "");
    mutation.mutate({
      object: "organization.workspace_quota",
      default_limit: defaultLimit,
      overrides: sanitized,
    });
  };

  const addOverride = () => {
    setOverrides((prev) => [
      ...prev,
      { user_public_id: "", limit: defaultLimit },
    ]);
  };

  const updateOverride = (index: number, value: Partial<OverrideState>) => {
    setOverrides((prev) => {
      const next = [...prev];
      next[index] = { ...next[index], ...value };
      return next;
    });
  };

  const removeOverride = (index: number) => {
    setOverrides((prev) => prev.filter((_, idx) => idx !== index));
  };

  return (
    <section className="mx-auto flex max-w-4xl flex-col gap-6 px-6 py-12">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold text-foreground">
          Workspace quotas
        </h1>
        <p className="text-sm text-muted-foreground">
          Define how many workspaces users can create. Overrides take precedence
          over the organization default limit.
        </p>
      </header>

      <Card>
        <CardHeader>
          <CardTitle>Limits</CardTitle>
          <CardDescription>
            Adjust the global default and optionally add user-specific
            overrides for special cases.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-6" onSubmit={handleSubmit}>
            <div className="space-y-2">
              <Label htmlFor="default-limit">Default workspace limit</Label>
              <Input
                id="default-limit"
                type="number"
                value={defaultLimit}
                onChange={(event) =>
                  setDefaultLimit(Math.max(1, Number(event.target.value)))
                }
                min={1}
                disabled={mutation.isPending || quotaQuery.isLoading}
              />
              <p className="text-xs text-muted-foreground">
                Applies to all members unless overridden below.
              </p>
            </div>

            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-sm font-semibold text-foreground">
                    Overrides
                  </h2>
                  <p className="text-xs text-muted-foreground">
                    Assign specific limits for selected users.
                  </p>
                </div>
                <Button
                  type="button"
                  variant="secondary"
                  onClick={addOverride}
                  disabled={mutation.isPending}
                >
                  Add override
                </Button>
              </div>

              {overrides.length === 0 && (
                <p className="text-sm text-muted-foreground">
                  No overrides configured yet.
                </p>
              )}

              <div className="space-y-3">
                {overrides.map((override, index) => (
                  <div
                    key={`${override.user_public_id}-${index}`}
                    className="grid gap-3 rounded-lg border border-border p-4 sm:grid-cols-[1fr_minmax(100px,160px)_auto]"
                  >
                    <div className="space-y-2">
                      <Label>User public ID</Label>
                      <Input
                        value={override.user_public_id}
                        onChange={(event) =>
                          updateOverride(index, {
                            user_public_id: event.target.value,
                          })
                        }
                        placeholder="user_abc123"
                        disabled={mutation.isPending}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>Limit</Label>
                      <Input
                        type="number"
                        min={1}
                        value={override.limit}
                        onChange={(event) =>
                          updateOverride(index, {
                            limit: Math.max(1, Number(event.target.value)),
                          })
                        }
                        disabled={mutation.isPending}
                      />
                    </div>
                    <div className="flex items-end justify-end">
                      <Button
                        type="button"
                        variant="ghost"
                        onClick={() => removeOverride(index)}
                        disabled={mutation.isPending}
                      >
                        Remove
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            <div className="flex justify-end gap-3">
              <Button
                type="submit"
                disabled={mutation.isPending || quotaQuery.isLoading}
              >
                {mutation.isPending ? "Saving…" : "Save changes"}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </section>
  );
}
