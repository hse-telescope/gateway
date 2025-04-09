FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache make
RUN make build

FROM alpine:latest AS runner
WORKDIR /app
COPY --from=builder /app/bin/gateway ./bin/gateway
ENTRYPOINT ["./bin/gateway"]
