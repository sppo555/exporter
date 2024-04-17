FROM golang:1.21.9-alpine3.19 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go install sd-dummy-exporter

#FROM golang:1.19-alpine3.18
FROM gcr.io/distroless/static:latest

COPY --from=builder /go/bin/sd-dummy-exporter ./main
COPY mq_consumer.log /tmp/mq_consumer.log
CMD ["./main"]