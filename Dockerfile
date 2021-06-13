FROM caddy:alpine
COPY Caddyfile /etc/caddy/Caddyfile
COPY srv/ /data/ipv6test.io/