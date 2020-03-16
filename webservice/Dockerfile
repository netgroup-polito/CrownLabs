# Dockerfile to build and deploy in production mode the web service
### builder image builds both the k8s libraries and Reack front-end
FROM node:alpine as builder
ADD . /webservice
WORKDIR /webservice
RUN npm install --silent --no-fund
WORKDIR /webservice/website
RUN npm install --silent --no-fund
RUN npm run-script build

### image to export the service
FROM nginx
COPY --from=builder /webservice/website/dist /usr/share/nginx/html
