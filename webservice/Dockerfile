FROM node:alpine

ADD . /webservice

# install k8s client
WORKDIR /webservice
RUN npm install --silent --no-fund

# install React-Webpack app
WORKDIR /webservice/website
RUN npm install --silent --no-fund

EXPOSE 8000
CMD ["npm", "start"]