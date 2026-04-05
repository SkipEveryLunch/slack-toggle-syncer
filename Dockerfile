# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -o main .

# Run stage
FROM alpine:3.21
WORKDIR /app

# タイムゾーンを日本時間に設定
ADD https://github.com/golang/go/raw/master/lib/time/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
ENV TZ=Asia/Tokyo

RUN apk add --no-cache tzdata && \
    addgroup -g 1000 appuser && \
    adduser -D -s /bin/sh -u 1000 -G appuser appuser

COPY --from=builder /app/main .
COPY --from=builder /app/config/config.toml.tmpl ./config/config.toml.tmpl

USER appuser

CMD ["./main"]
