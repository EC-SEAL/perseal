FROM node:12.16.1-alpine AS builder

WORKDIR /app

COPY package.json .


COPY . .

RUN npm set progress=false

RUN npm install

RUN npm run build --prod --baseHref=/per/


FROM nginx:1.15.8-alpine

COPY --from=builder /app/dist/perseal-ui /usr/share/nginx/html
COPY ./nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
