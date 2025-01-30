# Use the official Debian-based Go image
FROM golang:1.23.2-bullseye

# Set the working directory inside the container
WORKDIR /app

# 1. Install necessary packages:
#    - curl and xz-utils to download and extract the FFmpeg tarball
#    - ca-certificates to handle secure HTTPS
RUN apt-get update && apt-get install -y --no-install-recommends python3 && rm -rf /var/lib/apt/lists/*

# 4. Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# 5. Copy the rest of your source code and build
COPY . .
RUN go build -o discordbot ./cmd/musicbot/main.go

# 6. Run your bot
CMD ["./discordbot"]