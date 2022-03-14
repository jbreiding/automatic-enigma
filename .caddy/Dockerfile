FROM caddy:alpine

COPY Caddyfile /etc/caddy/Caddyfile
COPY index.html /srv/
RUN caddy fmt -overwrite /etc/caddy/Caddyfile
