# Multi-stage: frontend build -> Go single binary -> minimal runtime.
FROM node:20-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

FROM golang:1.27-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
COPY --from=frontend /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /app/app .

FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=backend /app/app /app/app
ENV APP_PORT=1067
EXPOSE 1067
USER nonroot:nonroot
ENTRYPOINT ["/app/app"]
