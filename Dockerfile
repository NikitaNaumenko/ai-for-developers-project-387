FROM node:24-alpine AS frontend-build

WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.26.3-alpine AS backend-build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/calendar-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
ENV PORT=10000
ENV STATIC_DIR=/frontend/dist
COPY --from=backend-build /out/calendar-api /calendar-api
COPY --from=frontend-build /src/frontend/dist /frontend/dist
EXPOSE 10000
USER nonroot:nonroot
ENTRYPOINT ["/calendar-api"]
