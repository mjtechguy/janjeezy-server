import { z } from "zod";

export const LocalLoginSchema = z.object({
  email: z.string().trim().email("Enter a valid email address"),
  password: z
    .string()
    .min(12, "Password must be at least 12 characters")
    .max(128, "Password is too long"),
});

export type LocalLoginInput = z.infer<typeof LocalLoginSchema>;
