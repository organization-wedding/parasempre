import { z } from "zod";

export const relationshipSchema = z.enum(["P", "R"]);

export const guestSchema = z.object({
  id: z.number(),
  first_name: z.string(),
  last_name: z.string(),
  relationship: relationshipSchema,
  attending: z.boolean().nullable(),
  family_group: z.number().int(),
  created_by: z.string(),
  updated_by: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});

export const guestsSchema = z.array(guestSchema);

export const createGuestInputSchema = z.object({
  first_name: z.string().trim().min(1, "Nome e sobrenome são obrigatórios."),
  last_name: z.string().trim().min(1, "Nome e sobrenome são obrigatórios."),
  relationship: relationshipSchema,
  family_group: z.number().int().min(1).optional(),
  phone: z.string().optional(),
});

export const updateGuestInputSchema = z
  .object({
    first_name: z.string().trim().min(1).optional(),
    last_name: z.string().trim().min(1).optional(),
    relationship: relationshipSchema.optional(),
    attending: z.boolean().nullable().optional(),
    family_group: z.number().int().min(1).optional(),
  })
  .refine((data) => Object.keys(data).length > 0, {
    message: "Pelo menos um campo deve ser enviado.",
  });

export const importResultSchema = z.object({
  imported: z.number().int(),
  errors: z.array(z.string()),
  total: z.number().int(),
});

export type Guest = z.infer<typeof guestSchema>;
export type CreateGuestInput = z.infer<typeof createGuestInputSchema>;
export type UpdateGuestInput = z.infer<typeof updateGuestInputSchema>;
export type ImportResult = z.infer<typeof importResultSchema>;

export const paginatedGuestsSchema = z.object({
  data: z.array(guestSchema),
  page: z.number(),
  limit: z.number(),
  total: z.number(),
});

export type PagedGuestResponse = z.infer<typeof paginatedGuestsSchema>;

export const guestStatsSchema = z.object({
  total: z.number().int(),
  confirmed: z.number().int(),
  pending: z.number().int(),
  declined: z.number().int(),
});

export type GuestStats = z.infer<typeof guestStatsSchema>;

export const myFamilyResponseSchema = z.array(guestSchema);

export const batchConfirmInputSchema = z.object({
  guest_ids: z.array(z.number().int().positive()).min(1),
  attending: z.boolean(),
});

export const meResponseSchema = z.object({
  role: z.string(),
  guest_id: z.number().int().nullable().optional(),
  first_name: z.string().nullable().optional(),
  last_name: z.string().nullable().optional(),
  family_group: z.number().int().nullable().optional(),
});

export type MyFamilyResponse = z.infer<typeof myFamilyResponseSchema>;
export type BatchConfirmInput = z.infer<typeof batchConfirmInputSchema>;
export type MeResponse = z.infer<typeof meResponseSchema>;
