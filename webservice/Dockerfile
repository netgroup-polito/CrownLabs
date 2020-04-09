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
COPY --from=builder_webservice /webservice/dist /usr/share/nginx/html
COPY --from=builder_webservice /webservice/nginx-default.conf /etc/nginx/conf.d/default.conf
COPY --from=builder_webservice /webservice/error-pages /usr/share/nginx/html/error-pages
COPY docker-entrypoint.sh /
COPY generate_config_js.sh /
RUN chmod +x docker-entrypoint.sh generate_config_js.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
