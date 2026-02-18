import { z } from "zod";

export const relationshipSchema = z.enum(["P", "R"]);

export const guestSchema = z.object({
  id: z.number(),
  first_name: z.string(),
  last_name: z.string(),
  phone: z.string().nullable(),
  relationship: relationshipSchema,
  confirmed: z.boolean(),
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
  phone: z.string(),
  relationship: relationshipSchema,
  family_group: z.number().int().min(1, "Grupo familiar deve ser um número válido."),
});

export const updateGuestInputSchema = z
  .object({
    first_name: z.string().trim().min(1).optional(),
    last_name: z.string().trim().min(1).optional(),
    phone: z.string().optional(),
    relationship: relationshipSchema.optional(),
    confirmed: z.boolean().optional(),
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
