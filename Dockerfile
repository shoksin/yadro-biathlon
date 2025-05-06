FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY cmd/ ./cmd/.
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o biathlon-processor ./cmd/main.go

FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/biathlon-processor .

COPY internal/config/config.json ./internal/config/
COPY internal/config/events ./internal/config/

VOLUME ["/app/results", "/app/logs"]

ENTRYPOINT ["./biathlon-processor"]

CMD ["--config_file=./internal/config/config.json", "--events_file=./internal/config/events", "--result_file=./results/resultingTable"]