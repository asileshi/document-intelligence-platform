# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod ./

RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ingestion-worker ./cmd/ingestion-worker

FROM alpine:3.20

RUN adduser -D -g '' app

WORKDIR /app

COPY --from=build /out/ingestion-worker /app/ingestion-worker

USER app

EXPOSE 9001

ENV PORT=9001

ENTRYPOINT ["/app/ingestion-worker"]
