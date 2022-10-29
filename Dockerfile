ARG CADDY_IMAGE
FROM $CADDY_IMAGE
COPY ./srv /srv
COPY Caddyfile /etc/caddy/Caddyfile
RUN mkdir /caddy_config && chown -R 8879 /caddy_config