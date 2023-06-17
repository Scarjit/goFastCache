# Stage 1: Building
FROM golang:alpine as builder
WORKDIR /go/src/app

# Download necessary Go packages
RUN apk add --no-cache git

# Copy Go Mod and Sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the Docker container
COPY ./pkg ./pkg

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./pkg

# Stage 2: Setup runtime environment
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /go/src/app/main .

# Expose port 8080 to the outside world
EXPOSE 8080

CMD ["./main"]
