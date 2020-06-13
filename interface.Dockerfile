FROM node:alpine AS builder

WORKDIR /app

COPY ./interface/package*.json ./

RUN npm ci

COPY ./interface .

EXPOSE 3000

CMD ["npm", "run", "start:nodemon"]
