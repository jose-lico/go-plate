FROM golang:1.22 as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/main ./cmd/main/main.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/main /app

CMD ["./main"]