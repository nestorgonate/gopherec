FROM golang:1.26.2-alpine3.22 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
WORKDIR /app/cmd/api
RUN go build -o /app/server .

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/server .
CMD ["./server"]