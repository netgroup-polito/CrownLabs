# Dockerfile to build and deploy in production mode the web service

### builder image builds both the k8s libraries and Reack front-end
FROM node:alpine as builder
ENV OIDC_CLIENT_ID="k8s" \
    OIDC_PROVIDER_URL="https://auth.crown-labs.ipv6.polito.it:4443/auth/realms/crownlabs" \
    APISERVER_URL="https://apiserver.crown-labs.ipv6.polito.it" \
    #Change OIDC_REDIRECT_URI to => http://localhost:8000 if you want to host your website locally
    #or in a docker run locally with the port 80 mapped to your host port 8000
    OIDC_REDIRECT_URI="https://crownlabs.polito.it"

ADD . /webservice
WORKDIR /webservice
RUN npm install --silent --unsafe-perm
WORKDIR /webservice/website
RUN npm install --silent --unsafe-perm
RUN npm run build

### image to export the service
FROM nginx
COPY --from=builder /webservice/website/dist /usr/share/nginx/html
