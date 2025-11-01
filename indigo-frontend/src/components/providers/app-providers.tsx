"use client";

import {
  QueryClient,
  QueryClientProvider,
  QueryClientConfig,
} from "@tanstack/react-query";
import { ReactNode, useState } from "react";
import { Toaster } from "sonner";
import { SessionRefresher } from "@/components/providers/session-refresher";
import { ThemeProvider } from "@/components/providers/theme-provider";

const queryClientConfig: QueryClientConfig = {
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      staleTime: 30_000,
      retry: 1,
    },
    mutations: {
      retry: 1,
    },
  },
};

export function AppProviders({ children }: { children: ReactNode }) {
  const [client] = useState(() => new QueryClient(queryClientConfig));

  return (
    <QueryClientProvider client={client}>
      <ThemeProvider>
        {children}
        <SessionRefresher />
        <Toaster position="top-center" richColors />
      </ThemeProvider>
    </QueryClientProvider>
  );
}
