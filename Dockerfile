ARG CADDY_IMAGE
FROM $CADDY_IMAGE
COPY Caddyfile /etc/caddy/Caddyfile
COPY srv/ /usr/share/caddy/