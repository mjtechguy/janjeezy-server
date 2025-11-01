"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { PropsWithChildren, useMemo } from "react";
import type { CurrentUser } from "@/lib/server/current-user";
import { LogoutButton } from "@/components/layout/logout-button";
import { cn } from "@/lib/utils";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";

type NavItem = {
  label: string;
  href: string;
  description?: string;
};

const NAV_ITEMS: NavItem[] = [
  { label: "Overview", href: "/admin/overview" },
  { label: "Organization", href: "/admin/organization" },
  { label: "Projects", href: "/admin/projects" },
  { label: "Providers", href: "/admin/providers" },
  { label: "Users", href: "/admin/users" },
  { label: "Invites", href: "/admin/invites" },
  { label: "SMTP Settings", href: "/admin/settings/smtp" },
  { label: "Workspace Quotas", href: "/admin/settings/workspace-quotas" },
  { label: "Audit Logs", href: "/admin/audit-logs" },
  { label: "API Keys", href: "/admin/api-keys" },
  { label: "Conversations", href: "/admin/conversations" },
  { label: "Responses", href: "/admin/responses" },
  { label: "MCP Tools", href: "/admin/mcp" },
];

export function AdminShell({
  user,
  children,
}: PropsWithChildren<{ user: CurrentUser }>) {
  const pathname = usePathname();
  const activeItem = useMemo(
    () =>
      NAV_ITEMS.find(
        (item) =>
          pathname === item.href ||
          (item.href !== "/admin/overview" && pathname.startsWith(item.href))
      ),
    [pathname]
  );

  return (
    <div className="flex min-h-screen bg-background text-foreground">
      <aside className="sticky top-0 hidden h-screen w-72 flex-col border-r bg-card/60 px-6 py-8 backdrop-blur lg:flex">
        <div className="mb-8 space-y-4">
          <Badge variant="secondary" className="uppercase tracking-[0.45em]">
            Jan Admin
          </Badge>
          <div className="space-y-2">
            <h2 className="text-lg font-semibold tracking-tight text-foreground">
              Control Plane
            </h2>
            <p className="text-sm text-muted-foreground">
              Manage organizations, providers, authentication, and platform
              operations from a centralized workspace.
            </p>
          </div>
        </div>
        <nav className="flex flex-1 flex-col gap-1">
          {NAV_ITEMS.map((item) => {
            const active =
              pathname === item.href ||
              (item.href !== "/admin/overview" && pathname.startsWith(item.href));
            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "group flex items-center rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                  active
                    ? "bg-primary text-primary-foreground shadow"
                    : "text-muted-foreground hover:bg-muted hover:text-foreground"
                )}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>
        <div className="mt-auto space-y-1 rounded-lg border border-dashed border-border/60 bg-card/40 p-4 text-sm text-muted-foreground">
          <p className="text-xs uppercase tracking-[0.3em]">Signed in</p>
          <p className="text-sm font-semibold text-foreground">{user.name}</p>
          <p className="text-xs text-muted-foreground">{user.email}</p>
        </div>
      </aside>

      <div className="flex flex-1 flex-col">
        <header className="sticky top-0 z-10 flex items-center justify-between border-b bg-background/80 px-4 py-4 backdrop-blur supports-[backdrop-filter]:bg-background/60 sm:px-6">
          <div>
            <p className="text-xs font-medium uppercase tracking-[0.35em] text-muted-foreground">
              Admin Console
            </p>
            <h1 className="text-xl font-semibold tracking-tight">
              {activeItem?.label ?? "Overview"}
            </h1>
          </div>
          <div className="flex items-center gap-3">
            <ThemeToggle />
            <Separator orientation="vertical" className="hidden h-6 lg:block" />
            <LogoutButton />
          </div>
        </header>
        <main className="flex-1 bg-background px-4 py-6 sm:px-6">
          <div className="mx-auto w-full max-w-6xl space-y-6">{children}</div>
        </main>
      </div>
    </div>
  );
}
