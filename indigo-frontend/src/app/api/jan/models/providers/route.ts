import { NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";

export const dynamic = "force-dynamic";

export async function GET() {
  try {
    const upstream = await janFetch("/v1/models/providers", {
      method: "GET",
    });
    const json = await upstream.json().catch(() => null);
    return NextResponse.json(json, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to load providers" },
      { status: 500 }
    );
  }
}
