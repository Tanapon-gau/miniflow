FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY . .
RUN GOFLAGS=-mod=mod go build -o /worker ./cmd/worker

FROM alpine:3.19
COPY --from=builder /worker /worker
CMD ["/worker"]
