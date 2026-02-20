import { serve } from "bun";
import index from "./index.html";

const port = parseInt(process.env.FRONTEND_PORT || "3000");
const isProd = process.env.NODE_ENV === "production";

let server;

if (isProd) {
  server = serve({
    port,
    async fetch(req) {
      const url = new URL(req.url);
      const pathname = url.pathname === "/" ? "/index.html" : url.pathname;
      const file = Bun.file(`./dist${pathname}`);
      if (await file.exists()) {
        return new Response(file);
      }
      // SPA fallback: unknown routes return index.html
      return new Response(Bun.file("./dist/index.html"));
    },
  });
} else {
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
