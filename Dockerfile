# Build the app
FROM golang:1.17

WORKDIR /build

# Install dependencies
COPY * ./
RUN go mod download

# Build server
RUN go build -o /server

CMD ["/server"]