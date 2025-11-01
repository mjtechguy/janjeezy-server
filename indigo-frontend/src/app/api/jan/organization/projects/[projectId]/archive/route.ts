import { NextRequest, NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";

export const dynamic = "force-dynamic";

export async function POST(_req: NextRequest, context: any) {
  try {
    const upstream = await janFetch(
      `/v1/organization/projects/${context.params.projectId}/archive`,
      {
        method: "POST",
      }
    );
    const json = await upstream.json().catch(() => null);
    return NextResponse.json(json, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to archive project" },
      { status: 500 }
    );
  }
}
