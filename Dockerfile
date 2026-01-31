FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o coolbot-exporter ./cmd/coolbot-exporter

FROM alpine:3.20
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/coolbot-exporter .

EXPOSE 9120

ENTRYPOINT ["./coolbot-exporter"]
