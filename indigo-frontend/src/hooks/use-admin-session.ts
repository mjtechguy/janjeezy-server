"use client";

import { useQuery, QueryKey } from "@tanstack/react-query";

export type AdminSession = {
  object: string;
  id: string;
  email: string;
  name: string;
};

type QueryData = AdminSession;

const SESSION_QUERY_KEY: QueryKey = ["admin-session"];

export function useAdminSession() {
  return useQuery<QueryData, Error>({
    queryKey: SESSION_QUERY_KEY,
    queryFn: async () => {
      const res = await fetch("/api/jan/auth/me", { cache: "no-store" });
      if (res.status === 401) {
        throw new Error("unauthorized");
      }
      if (!res.ok) {
        throw new Error("failed");
      }
      return (await res.json()) as QueryData;
    },
    retry: false,
    staleTime: 60_000,
  });
}

export function invalidateAdminSession(queryClient: {
  invalidateQueries: (opts: { queryKey: QueryKey }) => void;
}) {
  queryClient.invalidateQueries({ queryKey: SESSION_QUERY_KEY });
}
