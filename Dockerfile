FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /action-summary ./cmd/action-summary

FROM alpine:3.23

COPY --from=builder /action-summary /action-summary

ENTRYPOINT ["/action-summary"]
