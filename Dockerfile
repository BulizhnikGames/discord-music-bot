FROM golang:1.23.2-bullseye

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends python3 && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o discordbot ./cmd/musicbot/main.go

CMD ["./discordbot"]