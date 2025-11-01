import { NextRequest, NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";
import { copySetCookiesFromUpstream } from "@/lib/http/cookies";
import { setAccessTokenCookie } from "@/lib/auth/cookies";

export const dynamic = "force-dynamic";

export async function POST(req: NextRequest) {
  const payload = await req.json().catch(() => null);
  if (!payload) {
    return NextResponse.json({ error: "Invalid payload" }, { status: 400 });
  }

  try {
    const upstream = await janFetch("/v1/auth/google/callback", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    if (!upstream.ok) {
      const errorBody = await upstream.json().catch(() => null);
      return NextResponse.json(
        { error: errorBody?.error ?? "Google authentication failed" },
        { status: upstream.status }
      );
    }

    const body = (await upstream.json().catch(() => ({}))) as {
      access_token?: string;
      expires_in?: number;
    };
    const response = NextResponse.json(body);
    if (body.access_token && body.expires_in) {
      setAccessTokenCookie(response, body.access_token, body.expires_in);
    }
    copySetCookiesFromUpstream(upstream, response);
    return response;
  } catch {
    return NextResponse.json(
      { error: "Unexpected error completing Google login" },
      { status: 500 }
    );
  }
}
