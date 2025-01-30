FROM golang:1.23.2-alpine

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY .env cmd internal tools ./

RUN go build -o discordbot cmd/musicbot/main.go

CMD [ "./discordbot" ]