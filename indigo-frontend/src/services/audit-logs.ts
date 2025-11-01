import { AuditLogListSchema } from "@/schemas/audit-log";

export async function fetchAuditLogs(params?: { limit?: number; after?: string }) {
  const search = new URLSearchParams();
  if (params?.limit) {
    search.set("limit", String(params.limit));
  }
  if (params?.after) {
    search.set("after", params.after);
  }
  const res = await fetch(
    `/api/jan/organization/audit-logs${search.toString() ? `?${search.toString()}` : ""}`,
    {
      method: "GET",
      cache: "no-store",
    }
  );
  if (!res.ok) {
    throw new Error("Failed to load audit logs");
  }
  const json = await res.json();
  return AuditLogListSchema.parse(json);
}
