# syntax=docker/dockerfile:1

FROM node:22-bookworm-slim AS build

WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:1.27-alpine
# Only substitute ${NGINX_API_UPSTREAM}; leave nginx $uri-style vars untouched
ENV NGINX_ENVSUBST_FILTER=^NGINX_
COPY nginx/default.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80
