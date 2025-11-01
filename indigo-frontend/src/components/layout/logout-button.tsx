"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";

export function LogoutButton() {
  const router = useRouter();
  const [pending, setPending] = useState(false);

  const handleLogout = async () => {
    try {
      setPending(true);
      const res = await fetch("/api/jan/auth/logout", {
        method: "POST",
        cache: "no-store",
      });
      if (!res.ok) {
        toast.error("Failed to sign out");
      } else {
        toast.success("Signed out");
      }
      router.replace("/admin/login");
      router.refresh();
    } catch {
      toast.error("Unexpected logout error");
    } finally {
      setPending(false);
    }
  };

  return <Button variant="outline" size="sm" onClick={handleLogout} disabled={pending}>Sign out</Button>;
}
