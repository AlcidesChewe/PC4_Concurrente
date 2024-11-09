FROM golang:1.20
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY client/*.go ./client/
COPY utils ./utils
COPY config ./config

RUN go build -o client ./client/main.go

CMD ["./client"]

