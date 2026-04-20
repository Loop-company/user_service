FROM golang:1.25-alpine AS builder
WORKDIR /src

RUN apk add --no-cache git gcc musl-dev postgresql-client

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/user-service ./cmd

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata postgresql-client

COPY --from=builder /out/user-service /app/main

COPY --from=builder /src/migrations ./migrations

EXPOSE 8081

CMD ["/app/main"]