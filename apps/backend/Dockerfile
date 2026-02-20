FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/api ./cmd/api

FROM alpine:3.21

WORKDIR /app

RUN adduser -D -u 10001 appuser

COPY --from=builder /bin/api /app/api

EXPOSE 8080

USER appuser

CMD ["/app/api"]
