import { NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";

export const dynamic = "force-dynamic";

export async function GET() {
  try {
    const upstream = await janFetch("/v1/auth/me", { method: "GET" });

    if (!upstream.ok) {
      return NextResponse.json({ error: "Unauthorized" }, { status: upstream.status });
    }
    const body = await upstream.json();
    return NextResponse.json(body);
  } catch {
    return NextResponse.json({ error: "Unable to load session" }, { status: 500 });
  }
}
