# Dockerfile to build and deploy in production mode the web service

### builder image builds both the k8s libraries and Reack front-end
FROM node:alpine as builder

ADD . /webservice
WORKDIR /webservice
RUN npm install --silent --unsafe-perm
WORKDIR /webservice/website
RUN npm install --silent --unsafe-perm
RUN npm run build

### image to export the service
FROM nginx
COPY --from=builder /webservice/website/dist /usr/share/nginx/html
COPY docker-entrypoint.sh /
COPY generate_config_js.sh /
RUN chmod +x docker-entrypoint.sh generate_config_js.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
