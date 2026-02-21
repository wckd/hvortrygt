FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /hvortrygt .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /hvortrygt /hvortrygt
EXPOSE 8080
USER nobody
ENTRYPOINT ["/hvortrygt"]
