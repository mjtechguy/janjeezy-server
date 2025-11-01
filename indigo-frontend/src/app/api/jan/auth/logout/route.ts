import { NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";
import { deleteAccessTokenCookie } from "@/lib/auth/cookies";

export const dynamic = "force-dynamic";

export async function POST() {
  try {
    const upstream = await janFetch("/v1/auth/logout", { method: "GET" });
    if (!upstream.ok) {
      // Even if backend fails, clear local cookie to avoid lock-in
      const response = NextResponse.json({ error: "Failed to revoke session" }, { status: 500 });
      deleteAccessTokenCookie(response);
      return response;
    }
    const response = NextResponse.json({ success: true });
    deleteAccessTokenCookie(response);
    return response;
  } catch {
    const response = NextResponse.json({ error: "Unexpected logout error" }, { status: 500 });
    deleteAccessTokenCookie(response);
    return response;
  }
}
