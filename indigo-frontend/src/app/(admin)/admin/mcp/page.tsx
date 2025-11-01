"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchMcpActivity } from "@/services/mcp";

export default function McpPage() {
  const activityQuery = useQuery({
    queryKey: ["mcp-activity"],
    queryFn: fetchMcpActivity,
  });

  return (
    <section className="mx-auto flex max-w-6xl flex-col gap-6 px-6 py-12">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold text-foreground">
          MCP Activity
        </h1>
        <p className="text-sm text-muted-foreground">
          Inspect recent Model Context Protocol requests routed through the
          gateway.
        </p>
      </header>

      <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
        <table className="min-w-full divide-y divide-border/60">
          <thead className="bg-muted text-left text-xs font-semibold uppercase tracking-[0.3em] text-muted-foreground">
            <tr>
              <th className="px-6 py-3">Method</th>
              <th className="px-6 py-3">Tool</th>
              <th className="px-6 py-3">Timestamp</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border/40 text-sm text-muted-foreground">
            {activityQuery.isLoading && (
              <tr>
                <td className="px-6 py-4" colSpan={3}>
                  Loading MCP activity…
                </td>
              </tr>
            )}
            {activityQuery.error && !activityQuery.isLoading && (
              <tr>
                <td className="px-6 py-4 text-danger" colSpan={3}>
                  Unable to load activity.
                </td>
              </tr>
            )}
            {activityQuery.data?.data.map((entry, index) => (
              <tr key={`${entry.object}-${index}`} className="hover:bg-muted/60">
                <td className="px-6 py-4 font-medium text-foreground">
                  {entry.method}
                </td>
                <td className="px-6 py-4">
                  {entry.tool ?? "—"}
                </td>
                <td className="px-6 py-4">
                  {new Date(entry.created_at * 1000).toLocaleString()}
                </td>
              </tr>
            ))}
            {!activityQuery.isLoading &&
              !activityQuery.error &&
              (activityQuery.data?.data.length ?? 0) === 0 && (
                <tr>
                  <td className="px-6 py-4" colSpan={3}>
                    No activity recorded yet.
                  </td>
                </tr>
              )}
          </tbody>
        </table>
      </div>
    </section>
  );
}
