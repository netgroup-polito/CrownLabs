FROM node:12-alpine

COPY package*.json ./
RUN npm install
COPY src/*.js ./src/
ENV NODE_EXTRA_CA_CERTS /var/run/secrets/kubernetes.io/serviceaccount/ca.crt

ENTRYPOINT [ "npm", "start" ]