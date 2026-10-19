# Reverse proxy with the frontend: builds web/ and serves it next to the API.
# The build context is the repository root (see compose.yaml).
FROM node:24-alpine AS web
WORKDIR /app
COPY web/package.json web/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY web ./
RUN npm run build

FROM nginx:1.30-alpine
COPY deploy/nginx/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=web /app/dist /usr/share/nginx/html
