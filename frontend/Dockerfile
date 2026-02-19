FROM oven/bun:latest

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

COPY . .

ARG API_BASE=http://localhost:8080
RUN sed -i "s|http://localhost:8080|${API_BASE}|" bunfig.toml

ENV NODE_ENV=production

CMD ["bun", "src/index.ts"]
