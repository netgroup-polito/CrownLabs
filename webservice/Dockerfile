# Dockerfile to build and deploy in production mode the web service

### builder image builds both the k8s libraries and Reack front-end
FROM node:alpine as builder
ENV OIDC_CLIENT_ID="k8s" \
    OIDC_PROVIDER_URL="https://auth.crown-labs.ipv6.polito.it/auth/realms/crownlabs" \
    #Change OIDC_REDIRECT_URI to => http://localhost:8000 if you want to host your website locally (with npm start => webpack-dev-server)
    #or in a docker run locally with the port 80 mapped to your host port 8000 (docker run -p 8000:80 <tag>)
    OIDC_REDIRECT_URI="https://crownlabs.polito.it" \
    #The secret of the OIDC_CLIENT_ID
    #Should not be pushed when the repo becomes PUBLIC
    OIDC_CLIENT_SECRET="229a9d87-2bae-4e9b-8567-e8864b2bac4b" \
    APISERVER_URL="https://apiserver.crown-labs.ipv6.polito.it"

ADD . /webservice
WORKDIR /webservice
RUN npm install --silent --unsafe-perm
WORKDIR /webservice/website
RUN npm install --silent --unsafe-perm
RUN npm run build

### image to export the service
FROM nginx
COPY --from=builder /webservice/website/dist /usr/share/nginx/html

COPY --from=builder /webservice/nginx-default.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /webservice/website/error-pages /usr/share/nginx/html/error-pages