# Start from the official Golang image to ensure we have all the tools needed.
FROM golang:1.21.4

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first (for better cache utilization)
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application. Replace 'main.go' with the path to your application file if it's different.
RUN go build -o main .

# Expose port 9000 to the outside once the container has launched
EXPOSE 9000

# Run the executable
CMD ["./main"]
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