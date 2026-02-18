import { API_BASE } from "../config";
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

const RACF_KEY = "user-racf";

export function getUserRacf(): string | null {
  return localStorage.getItem(RACF_KEY);
}

export function setUserRacf(racf: string) {
  localStorage.setItem(RACF_KEY, racf.toUpperCase());
}

export function clearUserRacf() {
  localStorage.removeItem(RACF_KEY);
}

function authHeaders(): Record<string, string> {
  const racf = getUserRacf();
  if (!racf) throw new Error("Identificação (RACF) não configurada");
  return { "user-racf": racf };
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
  const data = await res.json();
  if (!res.ok) {
    throw new Error(typeof data?.error === "string" ? data.error : `Erro ${res.status}`);
  }
  return schema.parse(data);
}

export async function listGuests(): Promise<Guest[]> {
  const res = await fetch(`${API_BASE}/api/guests`);
  return handleResponse(res, guestsSchema);
}

export async function getGuest(id: number): Promise<Guest> {
  const res = await fetch(`${API_BASE}/api/guests/${id}`);
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
