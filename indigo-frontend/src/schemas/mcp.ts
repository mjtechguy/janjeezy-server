import { z } from "zod";

export const McpActivitySchema = z.object({
  object: z.string(),
  method: z.string(),
  tool: z.string().nullable().optional(),
  created_at: z.number(),
});

export type McpActivity = z.infer<typeof McpActivitySchema>;

export const McpActivityListSchema = z.object({
  object: z.literal("list"),
  data: z.array(McpActivitySchema),
  total: z.number().optional(),
});

export type McpActivityList = z.infer<typeof McpActivityListSchema>;
