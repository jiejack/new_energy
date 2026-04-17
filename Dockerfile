FROM golang:1.20-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o new-energy-monitoring ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/new-energy-monitoring .
COPY configs/ configs/

EXPOSE 8080

CMD ["./new-energy-monitoring"]
