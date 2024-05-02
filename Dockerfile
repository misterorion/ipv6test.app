FROM --platform=arm64 public.ecr.aws/docker/library/caddy:latest
COPY ./Caddyfile /etc/caddy/Caddyfile
COPY ./dist /srv