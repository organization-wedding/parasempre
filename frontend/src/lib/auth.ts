const TOKEN_KEY = "auth-token";
const ROLE_KEY = "auth-role";
const URACF_KEY = "auth-uracf";

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function getAuthRole(): string | null {
  return localStorage.getItem(ROLE_KEY);
}

export function getAuthUracf(): string | null {
  return localStorage.getItem(URACF_KEY);
}

export function isAuthenticated(): boolean {
  return getToken() !== null;
}

export function setAuth(token: string, role: string, uracf: string) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(ROLE_KEY, role);
  localStorage.setItem(URACF_KEY, uracf.toUpperCase());
  window.dispatchEvent(new Event("auth-changed"));
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(ROLE_KEY);
  localStorage.removeItem(URACF_KEY);
  window.dispatchEvent(new Event("auth-changed"));
}
