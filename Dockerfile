FROM golang:1.24.1-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod ./
COPY go.sum* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/server .

FROM alpine:3.18

RUN apk add --no-cache libc6-compat sqlite

WORKDIR /app

RUN mkdir -p /app/data

COPY --from=builder /app/server /app/

COPY --from=builder /app/templates* /app/templates/
COPY --from=builder /app/static* /app/static/

ENV GIN_MODE=release

EXPOSE 8080

CMD ["/app/server"]
