FROM golang:1.25 AS builder
WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o dndapi ./cmd


FROM alpine
WORKDIR /app
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/dndapi /app/dndapi
COPY migrations/ /app/migrations/
RUN ls -la /app/migrations/*
RUN touch .env


EXPOSE 8080
CMD ["/app/dndapi"]