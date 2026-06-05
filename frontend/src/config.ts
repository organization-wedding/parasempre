export const WEDDING_DATE = new Date("2026-10-12T16:00:00");

export const COUPLE = { name1: "Pedro Arthur", name2: "Rafaella Araujo" };

export const VENUE = {
  name: "Chácara Fênix",
  city: "Cambé, PR",
  address:
    "Av. Fabiano Dias Vector, 17 - Chácara Santa Maria, Cambé - PR, 86189-000",
  mapEmbed:
    "https://maps.google.com/maps?q=Chacara+Fenix,Cambe,PR,Brazil&t=&z=14&ie=UTF8&iwloc=&output=embed",
};

// Env vars come from src/_runtime-env.ts, regenerated from .env on every
// dev/build startup (see index.ts and build.ts). Single mechanism for dev
// and prod so they can't drift.
import { RUNTIME_ENV } from "./_runtime-env";

export const API_BASE: string = RUNTIME_ENV.API_BASE || "http://localhost:8080";
export const MERCADO_PAGO_PUBLIC_KEY: string = RUNTIME_ENV.MERCADO_PAGO_PUBLIC_KEY ?? "";

function apiHost(base: string): string {
  try {
    return new URL(base).hostname;
  } catch {
    return "";
  }
}

const API_HOST = apiHost(API_BASE);

export const IS_LOCAL_DEV = API_HOST === "localhost" || API_HOST === "127.0.0.1";
export const IS_DEV = IS_LOCAL_DEV || API_HOST.startsWith("teste.");

export const CONTACT = {
  phone: "(43) 99607-0599",
  phoneHref: "tel:+5543996070599",
  email: "pedroarthur1906@hotmail.com",
};

export const NAV_LINKS = [
  { label: "Registrar Presença", href: "/lista-presenca" },
  { label: "Lista de Presentes", href: "/lista-presentes" },
];
