import { z } from "zod";

export const SmtpSettingsSchema = z.object({
  object: z.literal("organization.smtp_settings"),
  enabled: z.boolean(),
  host: z.string(),
  port: z.number(),
  username: z.string(),
  from_email: z.string().email(),
  has_password: z.boolean().default(false),
});

export type SmtpSettings = z.infer<typeof SmtpSettingsSchema>;

export const WorkspaceQuotaOverrideSchema = z.object({
  user_public_id: z.string(),
  limit: z.number().int().positive(),
});

export const WorkspaceQuotaSchema = z.object({
  object: z.literal("organization.workspace_quota"),
  default_limit: z.number().int().positive(),
  overrides: z.array(WorkspaceQuotaOverrideSchema),
});

export type WorkspaceQuota = z.infer<typeof WorkspaceQuotaSchema>;
