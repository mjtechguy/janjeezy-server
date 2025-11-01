import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";
import { ACCESS_COOKIE_NAME } from "@/lib/constants/cookies";

const PUBLIC_ADMIN_ROUTES = ["/admin/login", "/admin/login/callback"];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const hasAccessToken = Boolean(request.cookies.get(ACCESS_COOKIE_NAME)?.value);

  const isPublic = PUBLIC_ADMIN_ROUTES.some((route) =>
    pathname.startsWith(route)
  );
  const isAdminRoute = pathname.startsWith("/admin");

  if (isAdminRoute && !isPublic && !hasAccessToken) {
    const url = request.nextUrl.clone();
    url.pathname = "/admin/login";
    url.searchParams.set("next", pathname);
    return NextResponse.redirect(url);
  }

  if (isPublic && hasAccessToken) {
    const url = request.nextUrl.clone();
    url.pathname = "/admin/overview";
    return NextResponse.redirect(url);
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/admin/:path*"],
};
