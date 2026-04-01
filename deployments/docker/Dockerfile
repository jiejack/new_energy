FROM golang:1.21-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git make

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api-server ./cmd/api-server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /collector ./cmd/collector
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /alarm ./cmd/alarm
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /compute ./cmd/compute
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ai-service ./cmd/ai-service
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /scheduler ./cmd/scheduler

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /api-server /app/
COPY --from=builder /collector /app/
COPY --from=builder /alarm /app/
COPY --from=builder /compute /app/
COPY --from=builder /ai-service /app/
COPY --from=builder /scheduler /app/
COPY --from=builder /build/configs /app/configs

ENV TZ=Asia/Shanghai
ENV GIN_MODE=release

EXPOSE 8080 9090

ENTRYPOINT ["/app/api-server"]
