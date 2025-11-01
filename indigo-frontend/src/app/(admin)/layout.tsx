import type { Metadata } from "next";
import { ReactNode } from "react";
import { AdminShell } from "@/components/layout/admin-shell";
import { getCurrentUser } from "@/lib/server/current-user";

export const metadata: Metadata = {
  title: "Jan Admin • Console",
};

export const dynamic = "force-dynamic";

export default async function AdminLayout({
  children,
}: {
  children: ReactNode;
}) {
  const user = await getCurrentUser();

  return <AdminShell user={user}>{children}</AdminShell>;
}
