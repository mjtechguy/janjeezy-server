import { NextRequest, NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";

export const dynamic = "force-dynamic";

export async function POST(req: NextRequest) {
  try {
    const segments = req.nextUrl.pathname.split("/");
    const providerId = segments[segments.length - 2];
    if (!providerId) {
      return NextResponse.json(
        { error: "Missing provider identifier" },
        { status: 400 }
      );
    }
    const upstream = await janFetch(
      `/v1/organization/models/providers/${providerId}/sync`,
      {
        method: "POST",
      }
    );
    const json = await upstream.json().catch(() => null);
    return NextResponse.json(json, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to sync provider models" },
      { status: 500 }
    );
  }
}
