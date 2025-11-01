import {
  OrganizationMemberList,
  OrganizationMemberListSchema,
} from "@/schemas/organization";

export async function fetchOrganizationMembers(): Promise<OrganizationMemberList> {
  const res = await fetch("/api/jan/organization/members", {
    method: "GET",
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error("Failed to load organization members");
  }
  const json = await res.json();
  return OrganizationMemberListSchema.parse(json);
}

export async function updateOrganizationMemberRole(
  userPublicId: string,
  role: "owner" | "reader"
) {
  const res = await fetch(`/api/jan/organization/members/${userPublicId}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ role }),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Failed to update member role");
  }
  return res.json();
}
