# Dockerfile to build and deploy in production mode the web service

# Builder image for the webservice
FROM node:14-alpine as builder

## Download git to fetch the kubernetes repo
RUN apk add --no-cache --update git openssh

## Switch to an unprivileged user (avoids problems with npm)
USER node

## Set the working directory and copy the source code
RUN mkdir --parent /tmp/webservice
COPY --chown=node:node . /tmp/webservice
WORKDIR /tmp/webservice

## Install the dependencies and build
RUN npm install
RUN npm run build


# Final image to export the service
FROM nginx:1.19

## Copy the different files
COPY --chown=nginx:nginx --from=builder /tmp/webservice/dist /usr/share/nginx/html
COPY --chown=nginx:nginx docker-entrypoint.sh /
COPY --chown=nginx:nginx generate_config_js.sh /
COPY --chown=nginx:nginx nginx_conf/default.conf /etc/nginx/conf.d/default.conf

## Add permissions for the nginx user
RUN chown -R nginx:nginx /usr/share/nginx/html
RUN chown -R nginx:nginx /var/cache/nginx && \
        chown -R nginx:nginx /var/log/nginx && \
        chown -R nginx:nginx /etc/nginx/conf.d
RUN touch /var/run/nginx.pid && \
        chown -R nginx:nginx /var/run/nginx.pid

## Allow nginx to bind to privileged ports even if the user is unprivileged
RUN apt update && apt install -y libcap2-bin
RUN setcap CAP_NET_BIND_SERVICE=+eip /usr/sbin/nginx

RUN chmod +x docker-entrypoint.sh generate_config_js.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
