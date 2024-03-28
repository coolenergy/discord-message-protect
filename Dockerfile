FROM golang:1.19.4 AS build

WORKDIR /app

COPY go.mod ./

# COPY *.go ./
COPY . ./

RUN go mod download
RUN go mod tidy

RUN go build -o /discord-message-protect

FROM debian:11.6-slim
RUN apt update && apt-get install -y ca-certificates

WORKDIR /app
COPY --from=build /discord-message-protect /app/discord-message-protect
COPY config.json /app/config.json

EXPOSE 9090

ENTRYPOINT /app/discord-message-protect run "$@"