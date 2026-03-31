# ─── Rust builder (with dependency caching) ──────────────────
FROM rust:1-slim AS api-build
RUN apt-get update && apt-get install -y pkg-config libssl-dev && rm -rf /var/lib/apt/lists/*
WORKDIR /app

# Copy manifests and create dummy source to cache dependency compilation
COPY api/Cargo.toml api/Cargo.lock ./api/
RUN mkdir -p api/src && echo 'fn main() {}' > api/src/main.rs
RUN cd api && cargo build --release

# Copy data directory (needed for include_str! at compile time)
COPY data ./data

# Now copy real source and rebuild (only app code recompiles)
COPY api/src ./api/src
RUN touch api/src/main.rs && cd api && cargo build --release

# ─── Node builder ────────────────────────────────────────────
FROM node:24-slim AS web-build
RUN corepack enable && corepack prepare pnpm@latest --activate
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ .
RUN pnpm build

# ─── API runtime ─────────────────────────────────────────────
FROM debian:bookworm-slim AS api
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=api-build /app/api/target/release/taxpilot-api .
COPY data ./data
EXPOSE 4100
CMD ["./taxpilot-api"]

# ─── Web runtime (static files + API reverse proxy via Caddy) ──
FROM caddy:2-alpine AS web
WORKDIR /app
COPY --from=web-build /app/web/dist ./dist
COPY web/Caddyfile /etc/caddy/Caddyfile
EXPOSE 4101
CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile"]
