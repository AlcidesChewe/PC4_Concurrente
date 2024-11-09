FROM golang:1.20
WORKDIR /app

COPY ../go.mod ../go.sum ./
RUN go mod download

COPY . .
COPY ../utils ../utils
COPY ../config ../config

RUN go build -o client main.go

CMD ["./client"]

