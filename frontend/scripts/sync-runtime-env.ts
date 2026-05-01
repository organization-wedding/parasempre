import { existsSync, readFileSync, writeFileSync } from "fs";
import { join } from "path";

// Generates src/_runtime-env.ts from the currently-loaded .env. Bun's HMR
// bundler doesn't substitute `process.env.X` in browser bundles, and HTML
// imports don't accept external script srcs — a regenerated TS file is the
// only reliable channel that works the same way in dev and prod.
//
// Idempotent: only writes when content changes, avoiding HMR reload loops.
export function syncRuntimeEnv(srcDir: string): void {
  const target = join(srcDir, "_runtime-env.ts");
  const next = `// AUTO-GENERATED on dev/build startup. Gitignored — regenerated from .env.
export const RUNTIME_ENV = {
  MERCADO_PAGO_PUBLIC_KEY: ${JSON.stringify(process.env.MERCADO_PAGO_PUBLIC_KEY ?? "")},
  API_BASE: ${JSON.stringify(process.env.API_BASE ?? "")},
} as const;
`;
  const current = existsSync(target) ? readFileSync(target, "utf8") : "";
  if (current !== next) writeFileSync(target, next);
}
