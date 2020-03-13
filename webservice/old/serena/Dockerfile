# build environment
FROM node:12.2.0-alpine as build
WORKDIR /frontend
ENV PATH /frontend/node_modules/.bin:$PATH
COPY package.json /frontend/package.json
RUN npm install --silent
RUN npm install react-scripts@3.0.1 -g --silent
COPY . /frontend
EXPOSE 3000
CMD ["npm", "start"]