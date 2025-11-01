import {
  CreateInviteSchema,
  InviteListSchema,
  InviteSchema,
} from "@/schemas/invite";

export async function fetchInvites() {
  const res = await fetch("/api/jan/organization/invites", {
    method: "GET",
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error("Failed to load invites");
  }
  const json = await res.json();
  return InviteListSchema.parse(json);
}

export async function createInvite(payload: {
  email: string;
  role: "owner" | "reader";
  projects: Array<{ id: string; role: "owner" | "member" }>;
}) {
  const parse = CreateInviteSchema.safeParse(payload);
  if (!parse.success) {
    throw new Error(parse.error.issues[0]?.message ?? "Invalid invite payload");
  }

  const res = await fetch("/api/jan/organization/invites", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(parse.data),
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Failed to create invite");
  }
  const json = await res.json();
  return InviteSchema.parse(json);
}

export async function deleteInvite(inviteId: string) {
  const res = await fetch(`/api/jan/organization/invites/${inviteId}`, {
    method: "DELETE",
  });
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? "Failed to delete invite");
  }
  return res.json();
}
