FROM node:12.2.0-alpine as build

# install k8s client
WORKDIR /webservice
ADD . /webservice
RUN npm install --silent

# install React-Webpack app
WORKDIR /webservice/website
COPY /website/package.json /webservice/website/package.json
RUN npm install --silent

EXPOSE 8000
CMD ["npm", "start"]