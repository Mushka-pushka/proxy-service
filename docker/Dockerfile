FROM golang:latest AS builder

RUN apt-get update && apt-get install -y git && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o proxy-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN adduser -D -g '' appuser

WORKDIR /root/

COPY --from=builder /app/proxy-service .

USER appuser

EXPOSE 8080

CMD ["./proxy-service"]