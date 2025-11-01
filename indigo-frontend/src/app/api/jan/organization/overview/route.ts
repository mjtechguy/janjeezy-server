import { NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";

export const dynamic = "force-dynamic";

export async function GET() {
  try {
    const upstream = await janFetch("/v1/organization/overview", {
      method: "GET",
    });
    const body = await upstream.json().catch(() => null);
    return NextResponse.json(body, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to load organization overview" },
      { status: 500 }
    );
  }
}
