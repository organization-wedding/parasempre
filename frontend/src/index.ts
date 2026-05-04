import { serve } from "bun";

const port = parseInt(process.env.FRONTEND_PORT || "3000");
const isProd = process.env.NODE_ENV === "production";

if (!isProd) {
  const { fileURLToPath } = await import("url");
  const { dirname } = await import("path");
  const { syncRuntimeEnv } = await import("../scripts/sync-runtime-env");
  syncRuntimeEnv(dirname(fileURLToPath(import.meta.url)));
}

function securityHeaders(): Record<string, string> {
  const apiBase = process.env.API_BASE || "";
  const csp = [
    "default-src 'self'",
    "script-src 'self' https://sdk.mercadopago.com",
    "script-src-elem 'self' https://sdk.mercadopago.com",
    "style-src 'self' 'unsafe-inline' https://fonts.googleapis.com",
    "font-src 'self' https://fonts.gstatic.com",
    "img-src 'self' data: https:",
    `connect-src 'self' https://api.mercadopago.com${apiBase ? " " + apiBase : ""}`,
    "frame-src https://*.mercadopago.com.br https://*.mercadopago.com",
    "frame-ancestors 'none'",
    "base-uri 'self'",
    "form-action 'self'",
    "object-src 'none'",
  ].join("; ");
  return {
    "Content-Security-Policy": csp,
    "X-Content-Type-Options": "nosniff",
    "Referrer-Policy": "strict-origin-when-cross-origin",
    "Strict-Transport-Security": "max-age=31536000; includeSubDomains",
  };
}

function withHeaders(res: Response, headers: Record<string, string>): Response {
  for (const [k, v] of Object.entries(headers)) res.headers.set(k, v);
  return res;
}

let server;

if (isProd) {
  const headers = securityHeaders();
  server = serve({
    port,
    async fetch(req) {
      const url = new URL(req.url);
      const pathname = url.pathname === "/" ? "/index.html" : url.pathname;
      const file = Bun.file(`./dist${pathname}`);
      if (await file.exists()) {
        return withHeaders(new Response(file), headers);
      }
      // SPA fallback: unknown routes return index.html
      return withHeaders(new Response(Bun.file("./dist/index.html")), headers);
    },
  });
} else {
  const { default: index } = await import("./index.html");
  server = serve({
    port,
    routes: {
      "/*": index,
    },
    development: {
      hmr: true,
      console: true,
    },
  });
}

console.log(`Server running at ${server.url}`);
