# Builder image for the frontend
FROM node:14-alpine as builder

## Switch to an unprivileged user (avoids problems with npm)
USER node

## Set the working directory and copy the source code
RUN mkdir --parent /tmp/frontend
WORKDIR /tmp/frontend

COPY --chown=node:node ./package.json /tmp/frontend/package.json
COPY --chown=node:node ./yarn.lock /tmp/frontend/yarn.lock
RUN yarn install

ARG BUILD_TARGET
ARG SUBROUTE="/"

ENV PUBLIC_URL=${SUBROUTE}

COPY --chown=node:node . /tmp/frontend/
RUN yarn build-${BUILD_TARGET}

# Final image to export the service
FROM nginx:1.19

ARG SUBROUTE="/"
ENV SUBROUTE=${SUBROUTE}

## Copy the different files
COPY --chown=nginx:nginx --from=builder /tmp/frontend/build /usr/share/nginx/html${SUBROUTE}
COPY --chown=nginx:nginx nginx.conf.tmpl /etc/nginx/conf.d/default.conf.tmpl

RUN envsubst '$SUBROUTE' < /etc/nginx/conf.d/default.conf.tmpl > /etc/nginx/conf.d/default.conf && \
    rm /etc/nginx/conf.d/default.conf.tmpl

## Add permissions for the nginx user
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chown -R nginx:nginx /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown -R nginx:nginx /var/run/nginx.pid

ENTRYPOINT ["nginx", "-g", "daemon off;"]
