FROM node:16.3-alpine

RUN mkdir --parent /qlkube
RUN chown node:node /qlkube && chmod 770 /qlkube

USER node
WORKDIR /qlkube

COPY --chown=node:node package.json ./
COPY --chown=node:node yarn.lock ./

RUN yarn install --production=true --frozen-lockfile

COPY src/*.js ./src/

ENV NODE_EXTRA_CA_CERTS /var/run/secrets/kubernetes.io/serviceaccount/ca.crt

ENTRYPOINT [ "yarn", "start" ]
