import { z } from "zod";

export const AuditLogSchema = z.object({
  object: z.literal("organization.audit_log"),
  id: z.number(),
  event: z.string(),
  user_email: z.string().nullish(),
  metadata: z.record(z.string(), z.any()),
  created_at: z.string(),
});

export type AuditLog = z.infer<typeof AuditLogSchema>;

export const AuditLogListSchema = z.object({
  object: z.literal("list"),
  data: z.array(AuditLogSchema),
  total: z.number(),
  first_id: z.string().nullable().optional(),
  last_id: z.string().nullable().optional(),
  has_more: z.boolean().optional(),
});

export type AuditLogList = z.infer<typeof AuditLogListSchema>;
