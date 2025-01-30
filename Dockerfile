FROM golang:1.23.2-alpine

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

RUN go build -o discordbot ./cmd/musicbot/main.go

CMD [ "./discordbot" ]