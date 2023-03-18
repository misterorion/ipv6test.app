FROM node:18-alpine AS builder
WORKDIR /app
COPY . .
RUN npm ci && npm run build

FROM busybox AS compressor
WORKDIR /app
COPY brotli.tar.gz .
COPY --from=builder /app/srv ./srv
RUN tar -zxf brotli.tar.gz && \
    find ./srv -type f -size +1400c \
    -regex ".*\.\(css\|js\|json\|svg\|xml\)$" \
    -exec ./brotli --best {} \+ \
    -exec gzip --best -k {} \+

FROM caddy:2-alpine
COPY --from=compressor /app/srv /srv
COPY ./Caddyfile /etc/caddy/Caddyfile
