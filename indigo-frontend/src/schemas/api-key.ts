import { z } from "zod";

export const AdminApiKeySchema = z.object({
  object: z.string(),
  id: z.string(),
  name: z.string().nullable().optional(),
  redacted_value: z.string().nullable().optional(),
  created_at: z.number(),
  last_used_at: z.number().nullable().optional(),
  owner: z.object({
    id: z.string().nullable().optional(),
    name: z.string().nullable().optional(),
    role: z.string().nullable().optional(),
  }).optional(),
  value: z.string().nullable().optional(),
});

export type AdminApiKey = z.infer<typeof AdminApiKeySchema>;

export const AdminApiKeyListSchema = z.object({
  object: z.literal("list"),
  data: z.array(AdminApiKeySchema),
  total: z.number().optional(),
});

export type AdminApiKeyList = z.infer<typeof AdminApiKeyListSchema>;
