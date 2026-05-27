# ── STAGE 1: Compilar ──
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY *.go ./
COPY go.mod ./
RUN CGO_ENABLED=0 GOOS=linux go build -o api .

# ── STAGE 2: Imagen final ──
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api .

EXPOSE 8080

CMD ["./api"]
