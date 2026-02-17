import { API_BASE } from "../config";
import type {
  Guest,
  CreateGuestInput,
  UpdateGuestInput,
  ImportResult,
} from "../types/guest";

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

async function handleResponse<T>(res: Response): Promise<T> {
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || `Erro ${res.status}`);
  }
  return data as T;
}

export async function listGuests(): Promise<Guest[]> {
  const res = await fetch(`${API_BASE}/api/guests`);
  return handleResponse<Guest[]>(res);
}

export async function getGuest(id: number): Promise<Guest> {
  const res = await fetch(`${API_BASE}/api/guests/${id}`);
  return handleResponse<Guest>(res);
}

export async function createGuest(input: CreateGuestInput): Promise<Guest> {
  const res = await fetch(`${API_BASE}/api/guests`, {
    method: "POST",
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify(input),
  });
  return handleResponse<Guest>(res);
}

export async function updateGuest(
  id: number,
  input: UpdateGuestInput,
): Promise<Guest> {
  const res = await fetch(`${API_BASE}/api/guests/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json", ...authHeaders() },
    body: JSON.stringify(input),
  });
  return handleResponse<Guest>(res);
}

export async function deleteGuest(id: number): Promise<void> {
  const res = await fetch(`${API_BASE}/api/guests/${id}`, {
    method: "DELETE",
    headers: authHeaders(),
  });
  if (!res.ok) {
    const data = await res.json();
    throw new Error(data.error || `Erro ${res.status}`);
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
  return handleResponse<ImportResult>(res);
}
