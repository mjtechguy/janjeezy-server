import { NextRequest, NextResponse } from "next/server";
import { janFetch } from "@/lib/jan-fetch";
import { LocalLoginSchema } from "@/schemas/auth";
import { copySetCookiesFromUpstream } from "@/lib/http/cookies";
import { setAccessTokenCookie } from "@/lib/auth/cookies";

export const dynamic = "force-dynamic";

export async function POST(req: NextRequest) {
  const raw = await req.json().catch(() => null);
  const parse = LocalLoginSchema.safeParse(raw);

  if (!parse.success) {
    return NextResponse.json(
      { error: parse.error.issues[0]?.message ?? "Invalid payload" },
      { status: 400 }
    );
  }

  try {
    const upstream = await janFetch("/v1/auth/local/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(parse.data),
    });

    if (!upstream.ok) {
      const payload = await upstream.json().catch(() => null);
      return NextResponse.json(
        { error: payload?.error ?? "Invalid email or password" },
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
      { error: "Unable to complete login. Please retry." },
      { status: 500 }
    );
  }
}
