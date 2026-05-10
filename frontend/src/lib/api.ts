import { API_BASE } from "../config";
import { getToken, clearAuth } from "./auth";
import type {
  Guest,
  CreateGuestInput,
  UpdateGuestInput,
  ImportResult,
  PagedGuestResponse,
} from "../types/guest";
import type {
  PublicGift,
  PagedGiftResponse,
  CreateGiftInput,
  UpdateGiftInput,
  CSVPreview,
  CommitImportResponse,
  ScrapePreviewResponse,
} from "../types/gift";
import {
  createGuestInputSchema,
  guestSchema,
  guestsSchema,
  importResultSchema,
  paginatedGuestsSchema,
  updateGuestInputSchema,
} from "../schemas/guest";
import {
  publicGiftSchema,
  paginatedGiftsSchema,
  createGiftInputSchema,
  updateGiftInputSchema,
  csvPreviewSchema,
  commitImportResponseSchema,
  scrapePreviewResponseSchema,
} from "../schemas/gift";
import {
  purchaseRequestSchema,
  purchaseResponseSchema,
  publicTransactionSchema,
  pagedPublicTransactionsSchema,
  pagedAdminTransactionsSchema,
  adminSummarySchema,
  type PurchaseRequest,
  type PurchaseResponse,
  type PublicTransaction,
  type AdminSummary,
} from "../schemas/payment";
import {
  publicMessageSchema,
  pagedPublicMessagesSchema,
  pagedAdminMessagesSchema,
  type PublicMessage,
  type PagedPublicMessages,
  type PagedAdminMessages,
} from "../schemas/giftMessage";

function authHeaders(): Record<string, string> {
  const token = getToken();
  if (!token) throw new Error("Autenticação necessária");
  return { Authorization: `Bearer ${token}` };
}

async function parseApiError(res: Response): Promise<string> {
  const contentType = res.headers.get("content-type") ?? "";
  if (contentType.includes("application/json")) {
    const body = await res.json();
    if (typeof body?.error === "string" && body.error.length > 0) {
      return body.error;
    }
  }
  return `Erro ${res.status}`;
}

export function isNotFoundError(error: unknown): boolean {
  return error instanceof Error && /not found|não encontrad/i.test(error.message);
}

async function handleResponse<T>(res: Response, schema: { parse: (value: unknown) => T }): Promise<T> {
  if (res.status === 401) {
    clearAuth();
    const path = encodeURIComponent(window.location.pathname);
    window.location.href = `/login?redirect=${path}`;
    throw new Error("Sessão expirada");
  }
  const data = await res.json();
  if (!res.ok) {
    throw new Error(typeof data?.error === "string" ? data.error : `Erro ${res.status}`);
  }
  return schema.parse(data);
}

export async function getUserMe(): Promise<{ role: string } | null> {
  const token = getToken();
  if (!token) return null;
  const res = await fetch(`${API_BASE}/api/users/me`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (res.status === 401) {
    clearAuth();
    return null;
  }
  if (res.status === 404) return null;
  if (!res.ok) {
    const data = (await res.json()) as { error?: string };
    throw new Error(typeof data?.error === "string" ? data.error : `Erro ${res.status}`);
  }
  return res.json() as Promise<{ role: string }>;
}

export async function listGuests(params?: { page?: number; limit?: number }): Promise<PagedGuestResponse> {
  const query = new URLSearchParams();
  if (params?.page) query.set("page", String(params.page));
  if (params?.limit) query.set("limit", String(params.limit));
  const url = query.toString() ? `${API_BASE}/api/guests?${query.toString()}` : `${API_BASE}/api/guests`;
  const res = await fetch(url, {
    headers: authHeaders(),
  });
  return handleResponse(res, paginatedGuestsSchema);
}

export async function getGuest(id: number): Promise<Guest> {
  const res = await fetch(`${API_BASE}/api/guests/${id}`, {
    headers: authHeaders(),
  });
  return handleResponse(res, guestSchema);
}

export async function createGuest(input: CreateGuestInput): Promise<Guest> {
  const payload = createGuestInputSchema.parse(input);
  const res = await fetch(`${API_BASE}/api/guests`, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify(payload),
  });
  return handleResponse(res, guestSchema);
}

export async function updateGuest(
  id: number,
  input: UpdateGuestInput,
): Promise<Guest> {
  const payload = updateGuestInputSchema.parse(input);
  const res = await fetch(`${API_BASE}/api/guests/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify(payload),
  });
  return handleResponse(res, guestSchema);
}

export async function deleteGuest(id: number): Promise<void> {
  const res = await fetch(`${API_BASE}/api/guests/${id}`, {
    method: "DELETE",
    headers: authHeaders(),
  });
  if (!res.ok) {
    throw new Error(await parseApiError(res));
  }
}

export async function importGuests(file: File): Promise<ImportResult> {
  const form = new FormData();
  form.append("file", file);
  const res = await fetch(`${API_BASE}/api/guests/import`, {
    method: "POST",
    headers: authHeaders(),
    body: form,
  });
  return handleResponse(res, importResultSchema);
}

// ── Gifts (pública) ─────────────────────────────────────────

export async function listGifts(params?: {
  page?: number;
  limit?: number;
}): Promise<PagedGiftResponse> {
  const query = new URLSearchParams();
  if (params?.page) query.set("page", String(params.page));
  if (params?.limit) query.set("limit", String(params.limit));
  const url = query.toString()
    ? `${API_BASE}/api/gifts?${query.toString()}`
    : `${API_BASE}/api/gifts`;
  const res = await fetch(url);
  return handleResponse(res, paginatedGiftsSchema);
}

export async function getGift(id: number): Promise<PublicGift> {
  const res = await fetch(`${API_BASE}/api/gifts/${id}`);
  return handleResponse(res, publicGiftSchema);
}

export async function createGift(input: CreateGiftInput): Promise<PublicGift> {
  const payload = createGiftInputSchema.parse(input);
  const res = await fetch(`${API_BASE}/api/gifts`, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify(payload),
  });
  return handleResponse(res, publicGiftSchema);
}

export async function updateGift(id: number, input: UpdateGiftInput): Promise<PublicGift> {
  const payload = updateGiftInputSchema.parse(input);
  const res = await fetch(`${API_BASE}/api/gifts/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify(payload),
  });
  return handleResponse(res, publicGiftSchema);
}

export async function deleteGift(id: number): Promise<void> {
  const res = await fetch(`${API_BASE}/api/gifts/${id}`, {
    method: "DELETE",
    headers: authHeaders(),
  });
  if (!res.ok) {
    throw new Error(await parseApiError(res));
  }
}

export async function previewGiftImport(file: File): Promise<CSVPreview> {
  const form = new FormData();
  form.append("file", file);
  const res = await fetch(`${API_BASE}/api/gifts/import/preview`, {
    method: "POST",
    headers: authHeaders(),
    body: form,
  });
  return handleResponse(res, csvPreviewSchema);
}

export async function commitGiftImport(
  rows: CreateGiftInput[],
): Promise<CommitImportResponse> {
  const res = await fetch(`${API_BASE}/api/gifts/import/commit`, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify({ rows }),
  });
  return handleResponse(res, commitImportResponseSchema);
}

export async function scrapeGiftURL(url: string): Promise<ScrapePreviewResponse> {
  const res = await fetch(`${API_BASE}/api/gifts/scrape-preview`, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify({ url }),
  });
  return handleResponse(res, scrapePreviewResponseSchema);
}

export async function createPurchase(
  giftId: number,
  body: PurchaseRequest,
): Promise<PurchaseResponse> {
  const payload = purchaseRequestSchema.parse(body);
  const res = await fetch(`${API_BASE}/api/gifts/${giftId}/purchase`, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify(payload),
  });
  return handleResponse(res, purchaseResponseSchema);
}

export async function listMyPurchases(params?: {
  page?: number;
  limit?: number;
}): Promise<ReturnType<typeof pagedPublicTransactionsSchema.parse>> {
  const query = new URLSearchParams();
  if (params?.page) query.set("page", String(params.page));
  if (params?.limit) query.set("limit", String(params.limit));
  const url = query.toString()
    ? `${API_BASE}/api/me/purchases?${query.toString()}`
    : `${API_BASE}/api/me/purchases`;
  const res = await fetch(url, { headers: authHeaders() });
  return handleResponse(res, pagedPublicTransactionsSchema);
}

export async function getMyPurchase(id: number): Promise<PublicTransaction> {
  const res = await fetch(`${API_BASE}/api/me/purchases/${id}`, {
    headers: authHeaders(),
  });
  return handleResponse(res, publicTransactionSchema);
}

export async function listAdminTransactions(filter?: {
  status?: string;
  gift_id?: number;
  page?: number;
  limit?: number;
}): Promise<ReturnType<typeof pagedAdminTransactionsSchema.parse>> {
  const query = new URLSearchParams();
  if (filter?.status) query.set("status", filter.status);
  if (filter?.gift_id) query.set("gift_id", String(filter.gift_id));
  if (filter?.page) query.set("page", String(filter.page));
  if (filter?.limit) query.set("limit", String(filter.limit));
  const url = query.toString()
    ? `${API_BASE}/api/transactions?${query.toString()}`
    : `${API_BASE}/api/transactions`;
  const res = await fetch(url, { headers: authHeaders() });
  return handleResponse(res, pagedAdminTransactionsSchema);
}

export async function adminTransactionsSummary(): Promise<AdminSummary> {
  const res = await fetch(`${API_BASE}/api/transactions/summary`, {
    headers: authHeaders(),
  });
  return handleResponse(res, adminSummarySchema);
}

// ── Gift Messages ───────────────────────────────────────────

export async function createGiftMessage(
  transactionId: number,
  formData: FormData,
): Promise<PublicMessage> {
  const res = await fetch(`${API_BASE}/api/transactions/${transactionId}/message`, {
    method: "POST",
    headers: authHeaders(),
    body: formData,
  });
  return handleResponse(res, publicMessageSchema);
}

export async function getMyTransactionMessage(
  transactionId: number,
): Promise<PublicMessage | null> {
  const res = await fetch(`${API_BASE}/api/transactions/${transactionId}/message`, {
    headers: authHeaders(),
  });
  if (res.status === 404) return null;
  return handleResponse(res, publicMessageSchema);
}

export async function listGiftMessages(
  giftId: number,
  params?: { page?: number; limit?: number },
): Promise<PagedPublicMessages> {
  const query = new URLSearchParams();
  if (params?.page) query.set("page", String(params.page));
  if (params?.limit) query.set("limit", String(params.limit));
  const url = query.toString()
    ? `${API_BASE}/api/gifts/${giftId}/messages?${query.toString()}`
    : `${API_BASE}/api/gifts/${giftId}/messages`;
  const res = await fetch(url);
  return handleResponse(res, pagedPublicMessagesSchema);
}

export async function listAdminGiftMessages(params?: {
  page?: number;
  limit?: number;
}): Promise<PagedAdminMessages> {
  const query = new URLSearchParams();
  if (params?.page) query.set("page", String(params.page));
  if (params?.limit) query.set("limit", String(params.limit));
  const url = query.toString()
    ? `${API_BASE}/api/admin/gift-messages?${query.toString()}`
    : `${API_BASE}/api/admin/gift-messages`;
  const res = await fetch(url, { headers: authHeaders() });
  return handleResponse(res, pagedAdminMessagesSchema);
}

export async function deleteGiftMessage(id: number): Promise<void> {
  const res = await fetch(`${API_BASE}/api/admin/gift-messages/${id}`, {
    method: "DELETE",
    headers: authHeaders(),
  });
  if (!res.ok && res.status !== 204) {
    throw new Error(await parseApiError(res));
  }
}

// ── Auth / OTP ──────────────────────────────────────────────

export class OtpApiError extends Error {
  readonly status: number;
  constructor(status: number, message: string) {
    super(message);
    this.name = "OtpApiError";
    this.status = status;
  }
}

export class OtpRateLimitError extends OtpApiError {
  readonly retryAfterSeconds: number;
  constructor(message: string, retryAfterSeconds: number) {
    super(429, message);
    this.name = "OtpRateLimitError";
    this.retryAfterSeconds = retryAfterSeconds;
  }
}

export async function sendOtp(phone: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/auth/otp/send`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ phone }),
  });
  if (res.ok) return;

  const data = (await res.json().catch(() => ({}))) as {
    error?: string;
    retry_after_seconds?: number;
  };
  const message = typeof data?.error === "string" ? data.error : `Erro ${res.status}`;
  if (res.status === 429 && typeof data?.retry_after_seconds === "number") {
    throw new OtpRateLimitError(message, data.retry_after_seconds);
  }
  throw new OtpApiError(res.status, message);
}

export interface TokenResponse {
  token: string;
  role: string;
  uracf: string;
}

export async function verifyOtp(phone: string, code: string): Promise<TokenResponse> {
  const res = await fetch(`${API_BASE}/api/auth/otp/verify`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ phone, code }),
  });
  if (!res.ok) {
    const data = (await res.json().catch(() => ({}))) as { error?: string };
    const message = typeof data?.error === "string" ? data.error : `Erro ${res.status}`;
    throw new OtpApiError(res.status, message);
  }
  return res.json() as Promise<TokenResponse>;
}

export async function devLogin(): Promise<TokenResponse> {
  const res = await fetch(`${API_BASE}/api/auth/dev-login`, { method: "POST" });
  if (!res.ok) {
    throw new Error(`dev-login failed: ${res.status}`);
  }
  return res.json() as Promise<TokenResponse>;
}
