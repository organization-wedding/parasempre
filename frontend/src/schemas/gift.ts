import { z } from "zod";

const httpsUrlSchema = z
  .string()
  .trim()
  .regex(/^https:\/\/[^\s]+$/, "URL inválida — use HTTPS.");

export const publicGiftSchema = z.object({
  id: z.number(),
  name: z.string(),
  description: z.string().nullish(),
  price_cents: z.number().int().nonnegative(),
  image_url: httpsUrlSchema.nullish(),
  store_url: httpsUrlSchema.nullish(),
  status: z.enum(["active", "inactive"]),
  created_at: z.string(),
  updated_at: z.string(),
});

export const paginatedGiftsSchema = z.object({
  data: z.array(publicGiftSchema),
  page: z.number(),
  limit: z.number(),
  total: z.number(),
});

export const createGiftInputSchema = z.object({
  name: z.string().trim().min(1, "Nome é obrigatório.").max(200, "Nome muito longo."),
  description: z.string().trim().max(2000, "Descrição muito longa.").optional(),
  price_cents: z.number().int().positive("Preço deve ser maior que zero."),
  image_url: httpsUrlSchema.optional(),
  store_url: httpsUrlSchema.optional(),
});

export const updateGiftInputSchema = createGiftInputSchema
  .partial()
  .refine((data) => Object.keys(data).length > 0, {
    message: "Pelo menos um campo deve ser alterado.",
  });

export type PublicGift = z.infer<typeof publicGiftSchema>;
export type PagedGiftResponse = z.infer<typeof paginatedGiftsSchema>;
export type CreateGiftInput = z.infer<typeof createGiftInputSchema>;
export type UpdateGiftInput = z.infer<typeof updateGiftInputSchema>;
