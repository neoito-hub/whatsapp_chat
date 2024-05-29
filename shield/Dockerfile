# Use the official Golang image as a base
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace
COPY . .

# Copy the go.mod and go.sum files to the working directory
COPY go.mod .
COPY go.sum .

# Download and install dependencies
RUN go mod download

#Build the Go application
RUN go build -o main .

# Expose the port your application runs on
EXPOSE 8080 8010 8011

# Command to run your application
CMD ["./main","packages.go"]