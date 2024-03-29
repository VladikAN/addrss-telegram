# Build image definition
FROM golang:1.20 AS builder
WORKDIR /src/
COPY . .
RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Runtime image definition
FROM alpine:latest
LABEL maintainer="https://github.com/VladikAN/addrss-telegram"
WORKDIR /root/

COPY --from=builder src/app .
COPY --from=builder src/templates/en/* templates/en/
COPY --from=builder src/templates/ru/* templates/ru/

ENTRYPOINT ["./app"]