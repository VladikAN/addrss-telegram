# Build image definition
FROM golang:1.14 AS builder
WORKDIR /src/
COPY . .
RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Runtime image definition
FROM alpine:latest
LABEL maintainer="https://github.com/VladikAN/feedreader-telegrambot"
WORKDIR /root/
COPY --from=builder src/app .
ENTRYPOINT ["./app"]