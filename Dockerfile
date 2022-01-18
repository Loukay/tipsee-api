FROM golang:1.17

WORKDIR /app

COPY . . 

# Install dependencies
RUN go mod download

# Build server
RUN go build -o /server

EXPOSE 3000

CMD [ "/server" ]