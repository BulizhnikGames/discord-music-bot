FROM golang:1.23.2-ubuntu

ARG LOGS_PATH

WORKDIR /app

RUN mkdir -p "$LOGS_PATH"

COPY go.* ./

RUN go mod download

COPY .env cmd internal tools ./

RUN go build -o discordbot cmd/musicbot/main.go

CMD [ "./discordbot" ]