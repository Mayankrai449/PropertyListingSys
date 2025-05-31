FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o property-listing-sys main.go

FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache ca-certificates


COPY --from=builder /app/property-listing-sys .

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./property-listing-sys"]