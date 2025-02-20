FROM golang:1.22.3-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod ./

COPY . .

RUN go mod download
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .

# Exponer el puerto
EXPOSE 8080

# Ejecutar la aplicación
CMD ["./server"]
