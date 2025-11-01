import { NextRequest, NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";

export const dynamic = "force-dynamic";

export async function DELETE(_req: NextRequest, context: any) {
  try {
    const { keyId } = context.params;
    const upstream = await janFetch(
      `/v1/organization/admin_api_keys/${keyId}`,
      {
        method: "DELETE",
      }
    );
    const json = await upstream.json().catch(() => null);
    return NextResponse.json(json, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to delete admin API key" },
      { status: 500 }
    );
  }
}
