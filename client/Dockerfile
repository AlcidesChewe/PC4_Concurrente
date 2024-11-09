FROM golang:1.20

# Set the working directory to /app
WORKDIR /app

# Copy go.mod and go.sum to /app
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the client source code into /app/client
COPY client/*.go ./client/

# Copy the utils and config packages
COPY utils ./utils
COPY config ./config

# Set the working directory to /app/client
WORKDIR /app/client

# Build the client binary
RUN go build -o client .

# Command to run when starting the container
CMD ["./client"]

