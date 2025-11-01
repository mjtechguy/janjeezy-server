import { z } from "zod";

export const ProviderSchema = z.object({
  id: z.string(),
  slug: z.string(),
  name: z.string(),
  vendor: z.string(),
  base_url: z.string().nullable().optional(),
  active: z.boolean(),
  metadata: z.record(z.string(), z.string()).optional(),
  scope: z.string(),
  project_id: z.string().nullable().optional(),
  last_synced_at: z.number().nullable().optional(),
  sync_latency_ms: z.number().nullable().optional(),
  api_key_hint: z.string().nullable().optional(),
  models_count: z.number().optional().default(0),
});

export type Provider = z.infer<typeof ProviderSchema>;

export const ProviderListSchema = z.object({
  object: z.literal("list"),
  data: z.array(ProviderSchema),
});

export type ProviderList = z.infer<typeof ProviderListSchema>;

export const ProviderVendorSchema = z.object({
  key: z.string(),
  name: z.string(),
  scope: z.string(),
  default_base_url: z.string().nullable().optional(),
  credential_hint: z.string().nullable().optional(),
});

export type ProviderVendor = z.infer<typeof ProviderVendorSchema>;

export const ProviderVendorListSchema = z.object({
  object: z.literal("list"),
  data: z.array(ProviderVendorSchema),
});

export type ProviderVendorList = z.infer<typeof ProviderVendorListSchema>;
