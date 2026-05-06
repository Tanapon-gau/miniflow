FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY . .
RUN GOFLAGS=-mod=mod go build -o /scheduler ./cmd/scheduler

FROM alpine:3.19
COPY --from=builder /scheduler /scheduler
CMD ["/scheduler"]
