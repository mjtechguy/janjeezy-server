import { McpActivityListSchema } from "@/schemas/mcp";

export async function fetchMcpActivity() {
  const res = await fetch("/api/jan/mcp/activity", {
    method: "GET",
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error("Failed to load MCP activity");
  }
  const json = await res.json();
  return McpActivityListSchema.parse(json);
}
