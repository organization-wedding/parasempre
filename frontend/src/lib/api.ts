import { API_BASE } from "../config";
import { getToken, clearAuth } from "./auth";
import type {
  Guest,
  CreateGuestInput,
  UpdateGuestInput,
  ImportResult,
} from "../types/guest";
import {
  createGuestInputSchema,
  guestSchema,
  guestsSchema,
  importResultSchema,
  updateGuestInputSchema,
} from "../schemas/guest";

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

export interface UserListItem {
  uracf: string;
  role: string;
  first_name: string;
  last_name: string;
}

export async function listUsers(): Promise<UserListItem[]> {
  const res = await fetch(`${API_BASE}/api/users`);
  if (!res.ok) {
    const data = (await res.json()) as { error?: string };
    throw new Error(typeof data?.error === "string" ? data.error : `Erro ${res.status}`);
  }
  return res.json() as Promise<UserListItem[]>;
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

export async function listGuests(): Promise<Guest[]> {
  const res = await fetch(`${API_BASE}/api/guests`, {
    headers: authHeaders(),
  });
  return handleResponse(res, guestsSchema);
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

// ── Auth / OTP ──────────────────────────────────────────────

export async function sendOtp(phone: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/auth/otp/send`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ phone }),
  });
  if (!res.ok) {
    const data = (await res.json()) as { error?: string };
    throw new Error(typeof data?.error === "string" ? data.error : `Erro ${res.status}`);
  }
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
    const data = (await res.json()) as { error?: string };
    throw new Error(typeof data?.error === "string" ? data.error : `Erro ${res.status}`);
  }
  return res.json() as Promise<TokenResponse>;
}
