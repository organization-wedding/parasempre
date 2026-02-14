
Default to using Bun instead of Node.js.

- Use `bun <file>` instead of `node <file>` or `ts-node <file>`
- Use `bun test` instead of `jest` or `vitest`
- Use `bun build <file.html|file.ts|file.css>` instead of `webpack` or `esbuild`
- Use `bun install` instead of `npm install` or `yarn install` or `pnpm install`
- Use `bun run <script>` instead of `npm run <script>` or `yarn run <script>` or `pnpm run <script>`
- Use `bunx <package> <command>` instead of `npx <package> <command>`
- Bun automatically loads .env, so don't use dotenv.

## APIs

- `Bun.serve()` supports WebSockets, HTTPS, and routes. Don't use `express`.
- `bun:sqlite` for SQLite. Don't use `better-sqlite3`.
- `Bun.redis` for Redis. Don't use `ioredis`.
- `Bun.sql` for Postgres. Don't use `pg` or `postgres.js`.
- `WebSocket` is built-in. Don't use `ws`.
- Prefer `Bun.file` over `node:fs`'s readFile/writeFile
- Bun.$`ls` instead of execa.

## Testing

Use `bun test` to run tests.

```ts#index.test.ts
import { test, expect } from "bun:test";

test("hello world", () => {
  expect(1).toBe(1);
});
```

## Frontend

Use HTML imports with `Bun.serve()`. Don't use `vite`. HTML imports fully support React, CSS, Tailwind.

Server:

```ts#index.ts
import index from "./index.html"

Bun.serve({
  routes: {
    "/": index,
    "/api/users/:id": {
      GET: (req) => {
        return new Response(JSON.stringify({ id: req.params.id }));
      },
    },
  },
  // optional websocket support
  websocket: {
    open: (ws) => {
      ws.send("Hello, world!");
    },
    message: (ws, message) => {
      ws.send(message);
    },
    close: (ws) => {
      // handle close
    }
  },
  development: {
    hmr: true,
    console: true,
  }
})
```

HTML files can import .tsx, .jsx or .js files directly and Bun's bundler will transpile & bundle automatically. `<link>` tags can point to stylesheets and Bun's CSS bundler will bundle.

```html#index.html
<html>
  <body>
    <h1>Hello, world!</h1>
    <script type="module" src="./frontend.tsx"></script>
  </body>
</html>
```

With the following `frontend.tsx`:

```tsx#frontend.tsx
import React from "react";
import { createRoot } from "react-dom/client";

// import .css files directly and it works
import './index.css';

const root = createRoot(document.body);

export default function Frontend() {
  return <h1>Hello, world!</h1>;
}

root.render(<Frontend />);
```

Then, run index.ts

```sh
bun --hot ./index.ts
```

For more information, read the Bun API docs in `node_modules/bun-types/docs/**.mdx`.

## Project Structure

```
src/
  config.ts              — Constants (dates, venue, contact, nav links)
  App.tsx                — Router (renders LandingPage or UnderConstruction)
  frontend.tsx           — React entry point
  index.css              — Minimal CSS (backgrounds, 3D transforms, animations)
  index.html             — HTML entry with favicon
  components/
    CoatOfArms.tsx       — Heraldic shield SVG
    OrnamentalDivider.tsx — Decorative divider
    Countdown.tsx        — Countdown timer
    Header.tsx           — Navigation + mobile menu
    Hero.tsx             — Hero section with frame
    MapSection.tsx       — Flip card with Google Maps
    Footer.tsx           — Site footer
  pages/
    LandingPage.tsx      — Assembles Header + Hero + MapSection + Footer
    UnderConstruction.tsx — WIP page with hourglass
  types/
    lucide-react.d.ts    — Type declarations for deep icon imports
```

## Code Conventions

- **Tailwind-first**: Use Tailwind classes for all styling. CSS only for complex backgrounds (gradients, SVG data URIs, noise textures), 3D transforms (`perspective`, `backface-visibility`), pseudo-elements (`::before`, `::after`), and keyframe animations.
- **Lucide icons**: Import from individual modules: `import MapPin from "lucide-react/dist/esm/icons/map-pin"` (not the barrel `"lucide-react"`).
- **Component per file**: Each component lives in its own file under `components/`. Pages under `pages/`.
- **State locality**: State lives in the component that owns it (e.g. `mapFlipped` in MapSection, `mobileMenuOpen` in Header).
- **Animation utilities**: Use CSS classes `anim-fade-in`, `anim-fade-in-slow`, `anim-fade-in-up`, `anim-fade-in-frame`, `anim-slide-down`, `anim-bounce` with `style={{ animationDelay }}` for staggered entrance animations.
- **Reusable backgrounds**: `dot-pattern`, `noise-texture`, `vignette` are generic CSS classes used across Hero and UnderConstruction pages. Opacity controlled via Tailwind (`opacity-15`, `opacity-[0.04]`).
- **Theme colors**: Defined in `styles/globals.css` under `@theme inline`. Use as Tailwind classes: `text-burgundy`, `bg-gold`, `border-gold-muted`, `font-heading`, `font-display`, etc.
