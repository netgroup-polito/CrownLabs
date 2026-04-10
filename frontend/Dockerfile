# Builder image for the frontend
FROM node:20 AS builder

## Switch to an unprivileged user (avoids problems with npm)
USER node

## Set the working directory and copy the source code
WORKDIR /tmp/frontend

COPY --chown=node:node ./package*.json .
RUN npm ci
COPY --chown=node:node . .

ARG BUILD_TARGET
ARG SUBROUTE="/"

ENV PUBLIC_URL=${SUBROUTE}

RUN npm run build-${BUILD_TARGET}

# Final image to export the service
FROM nginx:1.27

ARG SUBROUTE="/"
ENV SUBROUTE=${SUBROUTE}
ENV SUBROUTE_REDIR=${SUBROUTE%/}

## Copy the different files
COPY --chown=nginx:nginx --from=builder /tmp/frontend/dist /usr/share/nginx/html${SUBROUTE}
COPY --chown=nginx:nginx nginx.conf.tmpl /etc/nginx/conf.d/default.conf.tmpl

RUN envsubst '$SUBROUTE $SUBROUTE_REDIR' < /etc/nginx/conf.d/default.conf.tmpl > /etc/nginx/conf.d/default.conf && \
    rm /etc/nginx/conf.d/default.conf.tmpl

## Add permissions for the crownlabs user (uid/gid 20000)
RUN groupadd --gid 20000 crownlabs && \
    useradd --uid 20000 --gid 20000 --no-create-home --shell /usr/sbin/nologin crownlabs && \
    chown -R crownlabs:crownlabs /usr/share/nginx/html && \
    chown -R crownlabs:crownlabs /var/cache/nginx && \
    chown -R crownlabs:crownlabs /var/log/nginx && \
    chown -R crownlabs:crownlabs /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown crownlabs:crownlabs /var/run/nginx.pid

USER crownlabs

ENTRYPOINT ["nginx", "-g", "daemon off;"]
