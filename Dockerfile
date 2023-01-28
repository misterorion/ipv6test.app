FROM node:18-alpine AS builder
WORKDIR /app
COPY . .
RUN npm ci && npm run build

FROM caddy:2-alpine
COPY --from=builder /app/srv /srv
COPY ./Caddyfile /etc/caddy/Caddyfile
