import {
  SmtpSettingsSchema,
  WorkspaceQuotaSchema,
  WorkspaceQuota,
} from "@/schemas/settings";

export async function fetchSmtpSettings() {
  const res = await fetch("/api/jan/organization/settings/smtp", {
    method: "GET",
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error("Failed to load SMTP settings");
  }
  const json = await res.json();
  return SmtpSettingsSchema.parse(json);
}

export async function updateSmtpSettings(payload: {
  enabled: boolean;
  host: string;
  port: number;
  username: string;
  password?: string;
  from_email: string;
}) {
  const body: Record<string, unknown> = {
    enabled: payload.enabled,
    host: payload.host,
    port: payload.port,
    username: payload.username,
    from_email: payload.from_email,
  };
  if (payload.password !== undefined) {
    body.password = payload.password;
  }
  const res = await fetch("/api/jan/organization/settings/smtp", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const errorBody = await res.json().catch(() => null);
    throw new Error(errorBody?.error ?? "Failed to update SMTP settings");
  }
  const json = await res.json();
  return SmtpSettingsSchema.parse(json);
}

export async function fetchWorkspaceQuota() {
  const res = await fetch("/api/jan/organization/settings/workspace-quotas", {
    method: "GET",
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error("Failed to load workspace quotas");
  }
  const json = await res.json();
  return WorkspaceQuotaSchema.parse(json);
}

export async function updateWorkspaceQuota(payload: WorkspaceQuota) {
  const res = await fetch("/api/jan/organization/settings/workspace-quotas", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      default_limit: payload.default_limit,
      overrides: payload.overrides,
    }),
  });
  if (!res.ok) {
    const errorBody = await res.json().catch(() => null);
    throw new Error(errorBody?.error ?? "Failed to update workspace quotas");
  }
  const json = await res.json();
  return WorkspaceQuotaSchema.parse(json);
}
