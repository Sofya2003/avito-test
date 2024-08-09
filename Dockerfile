FROM golang:1.22.0 AS builder

WORKDIR /app

# Copy the source code
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Download and install the dependencies
RUN go get -d -v ./...

# Build the Go app
RUN go build -o main ./cmd/main.go
# RUN go test -o api ./internal/api

#WORKDIR /root/

# Копируем скомпилированное приложение из предыдущего этапа
#COPY --from=builder /app/main .

#EXPOSE the port
EXPOSE 8080

# Run the executable
CMD ["./main"]
