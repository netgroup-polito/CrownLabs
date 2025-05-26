FROM node:20-alpine

RUN mkdir --parent /qlkube
RUN chown node:node /qlkube && chmod 770 /qlkube

USER node
WORKDIR /qlkube

COPY --chown=node:node package*.json ./

RUN npm ci --omit=dev

COPY src/*.js ./src/

ENV NODE_EXTRA_CA_CERTS=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt

ENTRYPOINT [ "npm", "start" ]
