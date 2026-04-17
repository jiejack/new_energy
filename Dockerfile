FROM golang:1.21-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api-server ./cmd/api-server

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /api-server /app/
COPY --from=builder /build/configs /app/configs

ENV TZ=Asia/Shanghai
ENV GIN_MODE=release

EXPOSE 8080

ENTRYPOINT ["/app/api-server"]
