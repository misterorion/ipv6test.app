FROM caddy:2.4.3-alpine
COPY Caddyfile /etc/caddy/Caddyfile
COPY srv/ /usr/share/caddy/