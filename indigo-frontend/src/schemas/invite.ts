import { z } from "zod";

export const InviteProjectSchema = z.object({
  id: z.string(),
  role: z.string(),
});

export const InviteSchema = z.object({
  object: z.literal("organization.invite"),
  id: z.string(),
  email: z.string().email(),
  role: z.string(),
  status: z.string(),
  invited_at: z.string(),
  expires_at: z.string(),
  accepted_at: z.string().nullable().optional(),
  projects: z.array(InviteProjectSchema),
});

export const InviteListSchema = z.object({
  object: z.literal("list"),
  data: z.array(InviteSchema),
  total: z.number(),
  first_id: z.string().nullable().optional(),
  last_id: z.string().nullable().optional(),
  has_more: z.boolean().optional(),
});

export const CreateInviteSchema = z.object({
  email: z.string().email(),
  role: z.enum(["owner", "reader"]),
  projects: z
    .array(
      z.object({
        id: z.string(),
        role: z.enum(["owner", "member"]),
      })
    )
    .optional(),
});

export type Invite = z.infer<typeof InviteSchema>;
export type InviteList = z.infer<typeof InviteListSchema>;
export type CreateInviteInput = z.infer<typeof CreateInviteSchema>;
