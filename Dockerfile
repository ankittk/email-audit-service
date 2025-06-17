FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./main.go
RUN go build -o emailparser ./cmd/emailparser/main.go
RUN go build -o rulesengine ./cmd/rulesengine/main.go
RUN go build -o reportgenerator ./cmd/reportgenerator/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/emailparser .
COPY --from=builder /app/rulesengine .
COPY --from=builder /app/reportgenerator .

EXPOSE 8080 9090 9091 9092

CMD ["./main"]
