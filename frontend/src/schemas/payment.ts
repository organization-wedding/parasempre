import { z } from "zod";

const payerIdentificationSchema = z.object({
  type: z.literal("CPF"),
  number: z.string().trim().min(1),
});

export const payerSchema = z.object({
  email: z.string().email(),
  identification: payerIdentificationSchema,
});

export const purchaseRequestSchema = z.object({
  payment_method_id: z.string().min(1),
  token: z.string().min(1).optional(),
  issuer_id: z.string().optional(),
  installments: z.number().int().min(1).max(12).optional(),
  payer: payerSchema,
  idempotency_key: z.string().min(8).max(64),
});

export const pixDataSchema = z.object({
  qr_code: z.string().optional(),
  qr_code_base64: z.string().optional(),
  ticket_url: z.string().optional(),
});

export const purchaseResponseSchema = z.object({
  transaction_id: z.number().int(),
  mp_payment_id: z.string().nullish(),
  status: z.enum(["pending", "approved", "rejected", "refunded", "cancelled"]),
  status_detail: z.string().optional().default(""),
  payment_method: z.enum(["credit_card", "pix"]),
  amount_cents: z.number().int().nonnegative(),
  pix: pixDataSchema.nullish(),
});

export type PurchaseRequest = z.infer<typeof purchaseRequestSchema>;
export type PurchaseResponse = z.infer<typeof purchaseResponseSchema>;
export type PixData = z.infer<typeof pixDataSchema>;

export const publicTransactionSchema = z.object({
  id: z.number().int(),
  gift_id: z.number().int(),
  gift_name: z.string(),
  payment_method: z.enum(["credit_card", "pix"]),
  status: z.enum(["pending", "approved", "rejected", "refunded", "cancelled"]),
  amount_cents: z.number().int().nonnegative(),
  created_at: z.string(),
  updated_at: z.string(),
  pix: pixDataSchema.nullish(),
});

export const adminTransactionSchema = publicTransactionSchema.extend({
  user_id: z.number().int(),
  user_uracf: z.string(),
  user_phone: z.string().nullish(),
  mp_payment_id: z.string().nullish(),
});

export const pagedPublicTransactionsSchema = z.object({
  data: z.array(publicTransactionSchema),
  page: z.number().int(),
  limit: z.number().int(),
  total: z.number().int(),
});

export const pagedAdminTransactionsSchema = z.object({
  data: z.array(adminTransactionSchema),
  page: z.number().int(),
  limit: z.number().int(),
  total: z.number().int(),
});

export const statusBreakdownSchema = z.object({
  status: z.string(),
  count: z.number().int(),
  total_cents: z.number().int(),
});

export const adminSummarySchema = z.object({
  total: z.number().int(),
  approved_total_cents: z.number().int(),
  by_status: z.array(statusBreakdownSchema),
});

export type PublicTransaction = z.infer<typeof publicTransactionSchema>;
export type AdminTransaction = z.infer<typeof adminTransactionSchema>;
export type AdminSummary = z.infer<typeof adminSummarySchema>;
