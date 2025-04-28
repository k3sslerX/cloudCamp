FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY rate-limiting/ ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o rl ./main.go

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/rl .

RUN chmod +x rl

EXPOSE 8080
CMD ["/app/rl", "0.0.0.0:8080"]