"use client";

import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";

const REFRESH_INTERVAL_MS = 12 * 60 * 1000; // 12 minutes ~ <15m token lifetime

export function SessionRefresher() {
  const queryClient = useQueryClient();

  useEffect(() => {
    let active = true;

    async function refresh() {
      try {
        const res = await fetch("/api/jan/auth/refresh-token", {
          method: "GET",
          cache: "no-store",
        });
        if (res.ok) {
          queryClient.invalidateQueries({ queryKey: ["admin-session"] });
        }
      } catch {
        // ignore network errors; next cycle will retry
      }
    }

    const timer = setInterval(() => {
      if (!active) return;
      if (window.location.pathname.startsWith("/admin")) {
        void refresh();
      }
    }, REFRESH_INTERVAL_MS);

    return () => {
      active = false;
      clearInterval(timer);
    };
  }, [queryClient]);

  return null;
}
