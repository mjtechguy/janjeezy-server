"use client";

import { useMemo, useState } from "react";
import {
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import {
  createProvider,
  fetchProviderVendors,
  fetchProviders,
  updateProvider,
  syncProvider,
} from "@/services/providers";
import { fetchProjects } from "@/services/projects";
import { toast } from "sonner";
import type { ProviderVendor } from "@/schemas/provider";

export default function ProvidersPage() {
  const queryClient = useQueryClient();
  const providersQuery = useQuery({
    queryKey: ["providers"],
    queryFn: fetchProviders,
  });

  const vendorQuery = useQuery({
    queryKey: ["provider-vendors"],
    queryFn: fetchProviderVendors,
  });

  const projectsQuery = useQuery({
    queryKey: ["provider-projects"],
    queryFn: () => fetchProjects(),
  });

  const [form, setForm] = useState({
    name: "",
    vendor: "",
    baseURL: "",
    apiKey: "",
  });
  const [scope, setScope] = useState<"organization" | "project">(
    "organization"
  );
  const [projectId, setProjectId] = useState<string>("");
  const [errors, setErrors] = useState<{
    name?: string;
    vendor?: string;
    baseURL?: string;
    projectId?: string;
  }>({});
  const [baseUrlDirty, setBaseUrlDirty] = useState(false);
  const [autoBaseURL, setAutoBaseURL] = useState("");

  const createMutation = useMutation({
    mutationFn: createProvider,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["providers"] });
      toast.success("Provider registered");
      setForm({ name: "", vendor: "", baseURL: "", apiKey: "" });
      setScope("organization");
      setProjectId("");
      setErrors({});
      setBaseUrlDirty(false);
      setAutoBaseURL("");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const updateMutation = useMutation({
    mutationFn: ({
      id,
      payload,
    }: {
      id: string;
      payload: { active?: boolean };
    }) => updateProvider(id, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["providers"] });
      toast.success("Provider updated");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const syncMutation = useMutation({
    mutationFn: (providerId: string) => syncProvider(providerId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["providers"] });
      toast.success("Provider models synced");
    },
    onError: (err: Error) => toast.error(err.message),
  });

  const vendorOptions = useMemo<ProviderVendor[]>(
    () => vendorQuery.data?.data ?? [],
    [vendorQuery.data]
  );

  const selectedVendor = useMemo(
    () => vendorOptions.find((vendor) => vendor.key === form.vendor),
    [vendorOptions, form.vendor]
  );

  const handleVendorChange = (nextVendor: string) => {
    const meta = vendorOptions.find((vendor) => vendor.key === nextVendor);
    const defaultBase = meta?.default_base_url?.trim() ?? "";
    const currentTrimmed = form.baseURL.trim();
    const shouldApplyDefault =
      defaultBase !== "" &&
      (!baseUrlDirty || currentTrimmed === "" || currentTrimmed === autoBaseURL.trim());

    setForm((prev) => ({
      ...prev,
      vendor: nextVendor,
      baseURL: shouldApplyDefault ? defaultBase : prev.baseURL,
    }));
    setAutoBaseURL(defaultBase);
    setBaseUrlDirty((prevDirty) => (shouldApplyDefault ? false : prevDirty));
    setErrors((prev) => ({
      ...prev,
      vendor: undefined,
      baseURL: shouldApplyDefault ? undefined : prev.baseURL,
    }));
  };

  const handleRegister = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const nextErrors: typeof errors = {};
    const trimmedName = form.name.trim();
    const trimmedBaseURL = form.baseURL.trim();

    if (!trimmedName) {
      nextErrors.name = "Name is required";
    }
    if (!form.vendor) {
      nextErrors.vendor = "Select a vendor";
    }
    if (!trimmedBaseURL) {
      nextErrors.baseURL = "Base URL is required";
    } else {
      try {
        const parsed = new URL(trimmedBaseURL);
        if (!["http:", "https:"].includes(parsed.protocol)) {
          nextErrors.baseURL = "Base URL must start with http or https";
        }
      } catch {
        nextErrors.baseURL = "Enter a valid URL";
      }
    }
    if (scope === "project" && !projectId) {
      nextErrors.projectId = "Select a project";
    }

    setErrors(nextErrors);
    if (Object.keys(nextErrors).length > 0) {
      return;
    }

    createMutation.mutate({
      name: trimmedName,
      vendor: form.vendor,
      base_url: trimmedBaseURL,
      api_key: form.apiKey.trim() || undefined,
      project_public_id: scope === "project" ? projectId : undefined,
    });
  };

  return (
    <section className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-12">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold text-foreground">
          Model Providers
        </h1>
        <p className="text-sm text-muted-foreground">
          Connect upstream LLM providers and manage activation state for
          organization or project scopes.
        </p>
      </header>

      <form
        onSubmit={handleRegister}
        className="grid gap-4 rounded-2xl border border-border bg-card p-6 shadow-sm md:grid-cols-2"
      >
        <div className="md:col-span-1">
          <label className="text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
            Provider name
          </label>
          <input
            type="text"
            value={form.name}
            onChange={(event) => {
              const value = event.target.value;
              setForm((prev) => ({ ...prev, name: value }));
              setErrors((prev) => ({ ...prev, name: undefined }));
            }}
            placeholder="Production OpenAI"
            className={`mt-2 w-full rounded-xl border px-4 py-2 text-sm focus:outline-none ${
              errors.name
                ? "border-danger focus:border-danger"
                : "border-border focus-visible:ring-ring focus-visible:border-transparent"
            }`}
            disabled={createMutation.isPending}
          />
          {errors.name && (
            <p className="mt-2 text-xs text-danger">{errors.name}</p>
          )}
        </div>
        <div className="md:col-span-1">
          <label className="text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
            Vendor
          </label>
          <select
            value={form.vendor}
            onChange={(event) => handleVendorChange(event.target.value)}
            className={`mt-2 w-full rounded-xl border px-4 py-2 text-sm focus:outline-none ${
              errors.vendor
                ? "border-danger focus:border-danger"
                : "border-border focus-visible:ring-ring focus-visible:border-transparent"
            }`}
            disabled={createMutation.isPending || vendorQuery.isLoading}
          >
            <option value="">Select vendor</option>
            {vendorOptions.map((vendor) => (
              <option key={vendor.key} value={vendor.key}>
                {vendor.name}
              </option>
            ))}
          </select>
          {errors.vendor && (
            <p className="mt-2 text-xs text-danger">{errors.vendor}</p>
          )}
        </div>
        <div className="md:col-span-1">
          <label className="text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
            Scope
          </label>
          <select
            value={scope}
            onChange={(event) => {
              const next = event.target.value as "organization" | "project";
              setScope(next);
              if (next === "organization") {
                setProjectId("");
              }
              setErrors((prev) => ({ ...prev, projectId: undefined }));
            }}
            className="mt-2 w-full rounded-xl border border-border px-4 py-2 text-sm focus-visible:ring-ring focus-visible:border-transparent focus:outline-none"
            disabled={createMutation.isPending}
          >
            <option value="organization">Organization</option>
            <option value="project">Project</option>
          </select>
        </div>
        {scope === "project" && (
          <div className="md:col-span-1">
            <label className="text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
              Project
            </label>
            <select
              value={projectId}
              onChange={(event) => {
                setProjectId(event.target.value);
                setErrors((prev) => ({ ...prev, projectId: undefined }));
              }}
              className={`mt-2 w-full rounded-xl border px-4 py-2 text-sm focus:outline-none ${
                errors.projectId
                  ? "border-danger focus:border-danger"
                  : "border-border focus-visible:ring-ring focus-visible:border-transparent"
              }`}
              disabled={createMutation.isPending || projectsQuery.isLoading}
            >
              <option value="">Select project</option>
              {projectsQuery.data?.data.map((project) => (
                <option key={project.id} value={project.id}>
                  {project.name}
                </option>
              ))}
            </select>
            {errors.projectId && (
              <p className="mt-2 text-xs text-danger">{errors.projectId}</p>
            )}
          </div>
        )}
        <div className="md:col-span-1">
          <label className="text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
            Base URL
          </label>
          <input
            type="text"
            value={form.baseURL}
            onChange={(event) => {
              const value = event.target.value;
              setForm((prev) => ({ ...prev, baseURL: value }));
              setBaseUrlDirty(true);
              setErrors((prev) => ({ ...prev, baseURL: undefined }));
            }}
            placeholder={selectedVendor?.default_base_url ?? "https://api.openai.com/v1"}
            className={`mt-2 w-full rounded-xl border px-4 py-2 text-sm focus:outline-none ${
              errors.baseURL
                ? "border-danger focus:border-danger"
                : "border-border focus-visible:ring-ring focus-visible:border-transparent"
            }`}
            disabled={createMutation.isPending}
          />
          {errors.baseURL && (
            <p className="mt-2 text-xs text-danger">{errors.baseURL}</p>
          )}
          {selectedVendor?.default_base_url && (
            <p
              className={`${errors.baseURL ? "mt-1" : "mt-2"} text-xs text-muted-foreground`}
            >
              Recommended: {selectedVendor.default_base_url}
            </p>
          )}
        </div>
        <div className="md:col-span-1">
          <label className="text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
            API key (optional)
          </label>
          <input
            type="password"
            value={form.apiKey}
            onChange={(event) =>
              setForm((prev) => ({ ...prev, apiKey: event.target.value }))
            }
            placeholder={selectedVendor?.credential_hint ?? "sk-..."}
            className="mt-2 w-full rounded-xl border border-border px-4 py-2 text-sm focus-visible:ring-ring focus-visible:border-transparent focus:outline-none"
            disabled={createMutation.isPending}
          />
          {selectedVendor?.credential_hint && (
            <p className="mt-2 text-xs text-muted-foreground">
              Format: {selectedVendor.credential_hint}
            </p>
          )}
        </div>
        <div className="md:col-span-2">
          <button
            type="submit"
          className="h-10 rounded-xl bg-primary px-4 text-sm font-semibold text-primary-foreground transition hover:brightness-110 disabled:opacity-50"
            disabled={createMutation.isPending}
          >
            {createMutation.isPending ? "Registering…" : "Register provider"}
          </button>
        </div>
      </form>

      <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
        <table className="min-w-full divide-y divide-border/60">
          <thead className="bg-muted text-left text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
            <tr>
              <th className="px-6 py-3">Name</th>
              <th className="px-6 py-3">Vendor</th>
              <th className="px-6 py-3">Scope</th>
              <th className="px-6 py-3">Models</th>
              <th className="px-6 py-3">Last synced</th>
              <th className="px-6 py-3">Status</th>
              <th className="px-6 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40 text-sm">
            {providersQuery.isLoading && (
              <tr>
                <td className="px-6 py-4 text-muted-foreground" colSpan={7}>
                  Loading providers…
                </td>
              </tr>
            )}
            {providersQuery.error && !providersQuery.isLoading && (
              <tr>
                <td className="px-6 py-4 text-danger" colSpan={7}>
                  Failed to load providers.
                </td>
              </tr>
            )}
            {providersQuery.data?.data.map((provider) => (
              <tr key={provider.id} className="hover:bg-muted/60">
                <td className="px-6 py-4">
                  <div className="flex flex-col gap-1">
                    <span className="font-medium text-foreground">
                      {provider.name}
                    </span>
                    {provider.api_key_hint && (
                      <span className="text-xs text-muted-foreground">
                        Key hint: •••{provider.api_key_hint}
                      </span>
                    )}
                  </div>
                </td>
                <td className="px-6 py-4 text-muted-foreground">
                  {provider.vendor}
                </td>
                <td className="px-6 py-4 text-muted-foreground">
                  {provider.scope}
                </td>
                <td className="px-6 py-4 text-muted-foreground">
                  {provider.models_count ?? 0}
                </td>
                <td className="px-6 py-4 text-muted-foreground">
                  {provider.last_synced_at ? (
                    <div className="flex flex-col gap-1">
                      <span>
                        {new Date(provider.last_synced_at * 1000).toLocaleString()}
                      </span>
                      {provider.sync_latency_ms != null && (
                        <span className="text-xs text-muted-foreground">
                          Sync {provider.sync_latency_ms} ms
                        </span>
                      )}
                    </div>
                  ) : (
                    "—"
                  )}
                </td>
                <td className="px-6 py-4">
                  <span
                    className={`inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold ${
                      provider.active
                        ? "bg-success/10 text-success"
                        : "bg-warning/10 text-warning"
                    }`}
                  >
                    {provider.active ? "Active" : "Inactive"}
                  </span>
                </td>
                <td className="px-6 py-4 text-right">
                  <div className="flex items-center justify-end gap-2">
                    <button
                      type="button"
                      className="rounded-lg border border-border px-3 py-1 text-xs font-semibold text-muted-foreground transition hover:border-border hover:text-foreground disabled:opacity-50"
                      onClick={() => syncMutation.mutate(provider.id)}
                      disabled={syncMutation.isPending}
                    >
                      {syncMutation.isPending ? "Syncing…" : "Sync models"}
                    </button>
                    <button
                      type="button"
                      className="rounded-lg border border-border px-3 py-1 text-xs font-semibold text-muted-foreground transition hover:border-border hover:text-foreground disabled:opacity-50"
                      onClick={() =>
                        updateMutation.mutate({
                          id: provider.id,
                          payload: { active: !provider.active },
                        })
                      }
                      disabled={updateMutation.isPending}
                    >
                      {provider.active ? "Deactivate" : "Activate"}
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {!providersQuery.isLoading &&
              !providersQuery.error &&
              (providersQuery.data?.data.length ?? 0) === 0 && (
                <tr>
                  <td className="px-6 py-4 text-muted-foreground" colSpan={7}>
                    No providers registered yet.
                  </td>
                </tr>
              )}
          </tbody>
        </table>
      </div>
    </section>
  );
}
