FROM node:22-alpine AS frontend-build
WORKDIR /src/frontend
RUN corepack enable
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend ./
RUN pnpm build

FROM golang:1.23-alpine AS backend-build
WORKDIR /src
COPY backend/go.mod ./backend/
RUN cd backend && go mod download
COPY backend ./backend
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/hermes-web-studio ./backend/cmd/hermes-web-studio

FROM nginx:1.27-alpine
COPY deploy/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=frontend-build /src/frontend/dist /usr/share/nginx/html
COPY --from=backend-build /out/hermes-web-studio /usr/local/bin/hermes-web-studio
COPY deploy/start.sh /usr/local/bin/start-hermes-web-studio
RUN chmod +x /usr/local/bin/start-hermes-web-studio
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/start-hermes-web-studio"]
