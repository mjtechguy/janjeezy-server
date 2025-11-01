import { z } from "zod";

export const ProjectSchema = z.object({
  object: z.literal("project"),
  id: z.string(),
  name: z.string(),
  created_at: z.number(),
  archived_at: z.number().nullable().optional(),
  status: z.string(),
});

export type Project = z.infer<typeof ProjectSchema>;

export const ProjectListResponseSchema = z.object({
  object: z.literal("list"),
  data: z.array(ProjectSchema),
  first_id: z.string().nullable().optional(),
  last_id: z.string().nullable().optional(),
  has_more: z.boolean(),
});

export type ProjectListResponse = z.infer<typeof ProjectListResponseSchema>;
