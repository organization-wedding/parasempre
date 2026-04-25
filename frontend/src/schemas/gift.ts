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

const csvPreviewInputSchema = z.object({
  name: z.string(),
  description: z.string().nullish(),
  price_cents: z.number().int(),
  image_url: z.string().nullish(),
  store_url: z.string().nullish(),
  status: z.string().nullish(),
});

export const csvPreviewRowSchema = z.object({
  line_number: z.number().int(),
  status: z.enum(["new", "duplicate", "invalid"]),
  errors: z.array(z.string()).optional(),
  input: csvPreviewInputSchema,
  dedupe_key: z.string().optional(),
});

export const csvPreviewSchema = z.object({
  rows: z.array(csvPreviewRowSchema),
  summary: z.object({
    total: z.number().int(),
    new: z.number().int(),
    duplicate: z.number().int(),
    invalid: z.number().int(),
  }),
});

export const commitImportResponseSchema = z.object({
  created: z.number().int().nonnegative(),
  skipped: z.array(z.string()).optional(),
});

export const scrapePreviewResponseSchema = z.object({
  name: z.string(),
  description: z.string().optional().default(""),
  price_cents: z.number().int().nonnegative(),
  image_url: z.string().optional().default(""),
  store_url: z.string(),
});

export type PublicGift = z.infer<typeof publicGiftSchema>;
export type PagedGiftResponse = z.infer<typeof paginatedGiftsSchema>;
export type CreateGiftInput = z.infer<typeof createGiftInputSchema>;
export type UpdateGiftInput = z.infer<typeof updateGiftInputSchema>;
export type CSVPreviewRow = z.infer<typeof csvPreviewRowSchema>;
export type CSVPreview = z.infer<typeof csvPreviewSchema>;
export type CommitImportResponse = z.infer<typeof commitImportResponseSchema>;
export type ScrapePreviewResponse = z.infer<typeof scrapePreviewResponseSchema>;
