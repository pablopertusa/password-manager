# Usa la imagen oficial de Go para compilar el binario
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compila el binario estático, usa CGO_ENABLED=1 porque necesita interactuar con sqlite
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o server main.go

# Imagen Linux liviana
FROM alpine:latest

# Instala dependencias necesarias para SQLite
RUN apk update && apk add --no-cache ca-certificates sqlite

# Establece el directorio de trabajo
WORKDIR /root

# Copia el binario compilado desde la etapa anterior
COPY --from=builder /app/server /root/

EXPOSE 8080

CMD ["./server"]
