import { NextRequest, NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";

export const dynamic = "force-dynamic";

export async function PATCH(req: NextRequest, context: any) {
  const payload = await req.json().catch(() => null);
  try {
    const upstream = await janFetch(
      `/v1/organization/members/${context.params.userPublicId}`,
      {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      }
    );
    const json = await upstream.json().catch(() => null);
    return NextResponse.json(json, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to update organization member" },
      { status: 500 }
    );
  }
}
