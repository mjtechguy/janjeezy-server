import { NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";
import { copySetCookiesFromUpstream } from "@/lib/http/cookies";
import { setAccessTokenCookie } from "@/lib/auth/cookies";

export const dynamic = "force-dynamic";

export async function GET() {
  try {
    const upstream = await janFetch("/v1/auth/refresh-token", {
      method: "GET",
    });

    if (!upstream.ok) {
      return NextResponse.json({ error: "Unable to refresh" }, { status: upstream.status });
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
    return NextResponse.json({ error: "Unexpected refresh error" }, { status: 500 });
  }
}
