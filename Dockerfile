# Build stage
FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build -o server .

# Runtime stage
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/server .
COPY config.yaml .
EXPOSE 8080
CMD ["./server"]
