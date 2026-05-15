FROM golang:1.25.7-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

RUN adduser -D -u 1000 -g 1000 appuser

ENV TZ=Europe/Moscow

USER appuser

WORKDIR /home/appuser
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]