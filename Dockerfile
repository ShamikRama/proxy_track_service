FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata chromium

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata chromium

RUN adduser -D -s /bin/sh appuser

WORKDIR /app

COPY --from=builder /app/main .

COPY --from=builder /app/env.example .

RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]
