# Step 1
FROM golang:1.22-alpine AS builder

WORKDIR /app

# add server/go.sum when there is a go.sum file
COPY server/go.mod ./

RUN go mod download

COPY server/. .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Step 2
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"] 