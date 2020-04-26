# Dockerfile to build and deploy in production mode the web service

# builder image for the k8s_library js
FROM node:alpine as builder_k8s
WORKDIR /k8s_library
ADD ./k8s_library ./
RUN npm install --silent --unsafe-perm

# builder image for the webservice
FROM node:alpine as builder_webservice
WORKDIR /webservice
ADD . ./
COPY --from=builder_k8s /k8s_library ./k8s_library
RUN npm install --silent --unsafe-perm
RUN npm run build

### image to export the service
FROM nginx
COPY --chown=101:101 --from=builder_webservice /webservice/dist /usr/share/nginx/html
COPY --chown=101:101 docker-entrypoint.sh /
COPY --chown=101:101 generate_config_js.sh /
COPY --chown=101:101 nginx_conf/default.conf /etc/nginx/conf.d/default.conf
RUN chown -R 101:101 /usr/share/nginx/html
## add permissions for nginx user
RUN chown -R 101:101 /var/cache/nginx && \
        chown -R 101:101 /var/log/nginx && \
        chown -R 101:101 /etc/nginx/conf.d
RUN touch /var/run/nginx.pid && \
        chown -R 101:101 /var/run/nginx.pid


RUN chmod +x docker-entrypoint.sh generate_config_js.sh
RUN apt update && apt install -y libcap2-bin
RUN setcap CAP_NET_BIND_SERVICE=+eip /usr/sbin/nginx

ENTRYPOINT ["/docker-entrypoint.sh"]
