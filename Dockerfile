FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /shortener ./cmd/shortener

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /shortener /app/shortener

EXPOSE 8080

CMD ["/app/shortener"]
