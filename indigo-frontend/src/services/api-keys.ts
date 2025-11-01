import {
  AdminApiKeyListSchema,
  AdminApiKeySchema,
} from "@/schemas/api-key";

export async function fetchAdminApiKeys() {
  const res = await fetch("/api/jan/organization/admin-api-keys", {
    method: "GET",
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error("Failed to load API keys");
  }
  const json = await res.json();
  return AdminApiKeyListSchema.parse(json);
}

export async function createAdminApiKey(name: string) {
  const res = await fetch("/api/jan/organization/admin-api-keys", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Failed to create API key");
  }
  const json = await res.json();
  return AdminApiKeySchema.parse(json);
}

export async function deleteAdminApiKey(publicId: string) {
  const res = await fetch(
    `/api/jan/organization/admin-api-keys/${publicId}`,
    {
      method: "DELETE",
    }
  );
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Failed to delete API key");
  }
  return res.json();
}
