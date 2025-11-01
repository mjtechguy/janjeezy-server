import { NextRequest, NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";

export const dynamic = "force-dynamic";

export async function GET(req: NextRequest) {
  const search = req.nextUrl.search;
  try {
    const upstream = await janFetch(`/v1/organization/audit-logs${search}`, {
      method: "GET",
      cache: "no-store",
    });
    const json = await upstream.json().catch(() => null);
    return NextResponse.json(json, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to load audit logs" },
      { status: 500 }
    );
  }
}
