import { NextRequest, NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";

export const dynamic = "force-dynamic";

export async function GET(req: NextRequest) {
  const search = req.nextUrl.search;

  try {
    const upstream = await janFetch(`/v1/organization/projects${search}`, {
      method: "GET",
    });

    const body = await upstream.json().catch(() => null);

    return NextResponse.json(body, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to load organization projects" },
      { status: 500 }
    );
  }
}

export async function POST(req: NextRequest) {
  const body = await req.json().catch(() => null);
  try {
    const upstream = await janFetch(`/v1/organization/projects`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const payload = await upstream.json().catch(() => null);
    return NextResponse.json(payload, { status: upstream.status });
  } catch {
    return NextResponse.json(
      { error: "Unable to create project" },
      { status: 500 }
    );
  }
}
