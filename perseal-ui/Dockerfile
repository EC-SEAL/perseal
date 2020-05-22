
FROM node:12.16.1-alpine As builder

WORKDIR /app

COPY package.json .

RUN npm install

COPY . .

RUN npm run build --prod


FROM nginx:1.15.8-alpine

COPY --from=builder /app/dist/perseal-ui /usr/share/nginx/html
EXPOSE 80
