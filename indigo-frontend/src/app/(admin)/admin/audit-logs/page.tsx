"use client";

import { useMemo } from "react";
import { useInfiniteQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { fetchAuditLogs } from "@/services/audit-logs";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";

export default function AuditLogsPage() {
  const logsQuery = useInfiniteQuery({
    queryKey: ["audit-logs"],
    queryFn: ({ pageParam }) => fetchAuditLogs({ after: pageParam }),
    getNextPageParam: (lastPage) =>
      lastPage.has_more ? lastPage.last_id ?? undefined : undefined,
    initialPageParam: undefined as string | undefined,
    staleTime: 30_000,
    retry: 1,
  });

  const entries = useMemo(() => {
    return logsQuery.data?.pages.flatMap((page) => page.data) ?? [];
  }, [logsQuery.data]);

  const handleRefresh = async () => {
    try {
      await logsQuery.refetch();
      toast.success("Audit log refreshed");
    } catch (error) {
      console.error(error);
    }
  };

  return (
    <section className="mx-auto flex max-w-5xl flex-col gap-6 px-6 py-12">
      <header className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <h1 className="text-2xl font-semibold text-foreground">Audit logs</h1>
          <p className="text-sm text-muted-foreground">
            Review recent administrative actions captured by the API gateway.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            onClick={handleRefresh}
            disabled={logsQuery.isFetching}
          >
            {logsQuery.isFetching ? "Refreshing…" : "Refresh"}
          </Button>
        </div>
      </header>

      <Card>
        <CardHeader>
          <CardTitle>Activity</CardTitle>
          <CardDescription>
            Entries are ordered by most recent first and grouped per request.
          </CardDescription>
        </CardHeader>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-border/60 text-sm">
              <thead className="bg-muted text-left text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
                <tr>
                  <th className="px-6 py-3">Event</th>
                  <th className="px-6 py-3">Actor</th>
                  <th className="px-6 py-3">Metadata</th>
                  <th className="px-6 py-3">Created</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/40 text-muted-foreground">
                {logsQuery.isLoading && (
                  <tr>
                    <td className="px-6 py-4" colSpan={4}>
                      Loading audit entries…
                    </td>
                  </tr>
                )}
                {!logsQuery.isLoading && entries.length === 0 && (
                  <tr>
                    <td className="px-6 py-4" colSpan={4}>
                      No audit entries recorded yet.
                    </td>
                  </tr>
                )}
                {entries.map((entry) => (
                  <tr key={entry.id} className="hover:bg-muted/50">
                    <td className="px-6 py-4 font-medium text-foreground">
                      {entry.event}
                    </td>
                    <td className="px-6 py-4">
                      {entry.user_email ?? "system"}
                    </td>
                    <td className="px-6 py-4 text-xs">
                      <pre className="whitespace-pre-wrap break-words text-left text-xs text-muted-foreground">
                        {JSON.stringify(entry.metadata, null, 2)}
                      </pre>
                    </td>
                    <td className="px-6 py-4 text-sm">
                      {new Date(entry.created_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="flex justify-between border-t border-border/60 px-6 py-4">
            <span className="text-xs text-muted-foreground">
              Total entries: {logsQuery.data?.pages[0]?.total ?? entries.length}
            </span>
            <div className="flex items-center gap-3">
              <Button
                variant="secondary"
                onClick={() => logsQuery.fetchNextPage()}
                disabled={!logsQuery.hasNextPage || logsQuery.isFetchingNextPage}
              >
                {logsQuery.isFetchingNextPage
                  ? "Loading…"
                  : logsQuery.hasNextPage
                  ? "Load more"
                  : "No more"}
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </section>
  );
}
