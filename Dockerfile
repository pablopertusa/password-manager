FROM golang:1.23

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY utils/*.go ./utils/
COPY ./*.go .

# Build
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/server
COPY static /app/static
COPY .env /app/

COPY static /server/static

EXPOSE 2727

# Run
CMD ["/app/server"]