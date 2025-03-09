FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY utils/*.go ./utils/
COPY ./*.go .

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/server
COPY static /app/static
COPY .env /app/

COPY static /server/static

EXPOSE 2727

CMD ["/app/server"]