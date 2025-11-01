import {
  ProviderListSchema,
  ProviderVendorListSchema,
} from "@/schemas/provider";

export async function fetchProviders() {
  const res = await fetch("/api/jan/models/providers", {
    method: "GET",
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error("Failed to load providers");
  }
  const json = await res.json();
  return ProviderListSchema.parse(json);
}

export async function fetchProviderVendors() {
  const res = await fetch("/api/jan/organization/providers/vendors", {
    method: "GET",
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error("Failed to load provider vendors");
  }
  const json = await res.json();
  return ProviderVendorListSchema.parse(json);
}

export async function createProvider(payload: {
  name: string;
  vendor: string;
  base_url: string;
  api_key?: string;
  project_public_id?: string;
}) {
  const res = await fetch("/api/jan/organization/models/providers", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Failed to create provider");
  }
  return res.json();
}

export async function updateProvider(
  providerPublicId: string,
  payload: {
    name?: string;
    base_url?: string;
    api_key?: string;
    active?: boolean;
  }
) {
  const res = await fetch(
    `/api/jan/organization/models/providers/${providerPublicId}`,
    {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }
  );
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Failed to update provider");
  }
  return res.json();
}

export async function syncProvider(providerPublicId: string) {
  const res = await fetch(
    `/api/jan/organization/models/providers/${providerPublicId}/sync`,
    {
      method: "POST",
    }
  );
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Failed to sync provider models");
  }
  return res.json();
}
