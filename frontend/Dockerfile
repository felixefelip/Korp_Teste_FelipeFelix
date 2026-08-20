FROM node:24-slim AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

# ---------- dev ----------
# Usado pelo docker compose com bind mount: ng serve com hot reload.
FROM node:24-slim AS dev
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
EXPOSE 4200
CMD ["npm", "start", "--", "--host", "0.0.0.0"]

FROM deps AS build
COPY . .
RUN npm run build

# ---------- prod ----------
FROM nginx:alpine AS prod

ENV API_UPSTREAM=http://estoque:8000 \
    DNS_RESOLVER=127.0.0.11 \
    NGINX_ENVSUBST_FILTER='API_|DNS_'

COPY nginx.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist/frontend/browser /usr/share/nginx/html
EXPOSE 80
