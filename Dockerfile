FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server/main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/server .

EXPOSE 8080

CMD ["./server"]
