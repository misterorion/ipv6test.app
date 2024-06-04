FROM public.ecr.aws/docker/library/caddy:2.8.4
COPY ./Caddyfile /etc/caddy/Caddyfile
COPY ./dist /srv
