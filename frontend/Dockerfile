# Build stage: install deps and compile static assets
FROM oven/bun:latest AS builder

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

COPY . .

ARG API_BASE=http://localhost:8080
RUN API_BASE=${API_BASE} bun run build

# Production stage: serve pre-built dist/ as static files
FROM oven/bun:latest AS runner

WORKDIR /app

COPY --from=builder /app/dist ./dist
COPY --from=builder /app/src/index.ts ./src/index.ts

ENV NODE_ENV=test

CMD ["bun", "src/index.ts"]
