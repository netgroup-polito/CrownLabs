# Dockerfile to build and deploy in production mode the web service
### builder image builds both the k8s libraries and Reack front-end
FROM node as builder

RUN apt-get update
RUN apt-get install -y nginx

ADD . /webservice

WORKDIR /webservice
RUN npm install --silent --unsafe-perm
WORKDIR /webservice/website
RUN npm install --silent --unsafe-perm

COPY nginx.conf /etc/nginx/conf.d/default.conf

RUN chmod +x ../docker_entrypoint.sh

ENTRYPOINT ["../docker_entrypoint.sh"]
