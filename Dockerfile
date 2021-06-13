FROM caddy:2.4.1-alpine
COPY Caddyfile /etc/caddy/Caddyfile
COPY srv/ /usr/share/caddy/