import { NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";

export const dynamic = "force-dynamic";

export async function GET() {
  try {
    const res = await janFetch("/v1/auth/google/login", { method: "GET" });
    if (!res.ok) {
      return NextResponse.json(
        { error: "Unable to initialise Google login" },
        { status: 502 }
      );
    }
    const body = await res.json();
    return NextResponse.json({ redirectUrl: body.url });
  } catch {
    return NextResponse.json(
      { error: "Unexpected error starting Google login" },
      { status: 500 }
    );
  }
}
