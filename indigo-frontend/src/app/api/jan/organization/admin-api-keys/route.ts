import { NextRequest, NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";

export const dynamic = "force-dynamic";

export async function GET(req: NextRequest) {
  const search = req.nextUrl.search;
  try {
    const upstream = await janFetch(`/v1/organization/admin_api_keys${search}`, {
      method: "GET",
    });
    const json = await upstream.json().catch(() => null);
    return NextResponse.json(json, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to load admin API keys" },
      { status: 500 }
    );
  }
}

export async function POST(req: NextRequest) {
  const payload = await req.json().catch(() => null);
  try {
    const upstream = await janFetch("/v1/organization/admin_api_keys", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const json = await upstream.json().catch(() => null);
    return NextResponse.json(json, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to create admin API key" },
      { status: 500 }
    );
  }
}
