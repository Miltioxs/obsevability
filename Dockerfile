FROM golang:1.18.0-alpine AS builder

WORKDIR /app
COPY . .

RUN apk update && apk add --no-cache git
RUN CGO_ENABLED=0 GOOS=linux go build -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE 7007
CMD ["./app"]