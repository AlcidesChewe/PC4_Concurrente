FROM golang:1.20
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY server/*.go ./server/
COPY utils ./utils
COPY config ./config
COPY data ./data

RUN go build -o server ./server/main.go

EXPOSE 8000

CMD ["./server"]

