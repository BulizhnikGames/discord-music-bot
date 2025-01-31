FROM golang:1.23.2-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o discordbot ./cmd/musicbot/main.go

FROM debian:bullseye-slim

WORKDIR /app

RUN /bin/sh -c set -eux && \
    apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates python3 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /app ./

RUN rm -rf cmd internal go.mod go.sum

CMD ["./discordbot"]