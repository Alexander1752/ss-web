FROM node as builder

WORKDIR /app

COPY package*.json ./

RUN yarn install --frozen-lockfile

COPY . .

RUN yarn build


FROM nginx:stable-alpine

COPY --from=builder /app/dist/ /usr/share/nginx/html/
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
