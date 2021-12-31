ARG CADDY_IMAGE
FROM $CADDY_IMAGE
COPY Caddyfile /etc/caddy/Caddyfile
COPY srv/ /usr/share/caddy/
RUN mkdir /caddy_config && chown -R 8879 /caddy_config