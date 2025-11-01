import { NextRequest, NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";

export const dynamic = "force-dynamic";

export async function GET() {
  try {
    const upstream = await janFetch("/v1/organization/settings/smtp", {
      method: "GET",
      cache: "no-store",
    });
    const body = await upstream.json().catch(() => null);
    return NextResponse.json(body, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to load SMTP settings" },
      { status: 500 }
    );
  }
}

export async function PUT(req: NextRequest) {
  const payload = await req.json().catch(() => null);
  try {
    const upstream = await janFetch("/v1/organization/settings/smtp", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const json = await upstream.json().catch(() => null);
    return NextResponse.json(json, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to update SMTP settings" },
      { status: 500 }
    );
  }
}
