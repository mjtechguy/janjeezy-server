import { z } from "zod";

export const OrganizationOverviewSchema = z.object({
  object: z.literal("organization.overview"),
  projects: z.object({
    total: z.number(),
    active: z.number(),
    archived: z.number(),
  }),
  members: z.object({
    total: z.number(),
  }),
  invites: z.object({
    pending: z.number(),
  }),
  providers: z.object({
    active: z.number(),
    inactive: z.number(),
  }),
});

export type OrganizationOverview = z.infer<typeof OrganizationOverviewSchema>;

export const OrganizationMemberSchema = z.object({
  object: z.literal("organization.member"),
  role: z.string(),
  joined_at: z.number(),
  user: z.object({
    id: z.string(),
    name: z.string(),
    email: z.string(),
    created_at: z.number(),
  }),
});

export type OrganizationMember = z.infer<typeof OrganizationMemberSchema>;

export const OrganizationMemberListSchema = z.object({
  object: z.literal("list"),
  data: z.array(OrganizationMemberSchema),
  total: z.number(),
  first_id: z.string().nullable().optional(),
  last_id: z.string().nullable().optional(),
  has_more: z.boolean().optional(),
});

export type OrganizationMemberList = z.infer<
  typeof OrganizationMemberListSchema
>;
