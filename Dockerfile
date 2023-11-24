FROM golang:alpine as builder

WORKDIR /build

COPY main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -o deeplx-load-balancer main.go

FROM alpine:latest as prod

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=0 /build/deeplx-load-balancer .

EXPOSE 8080
CMD ["./deeplx-load-balancer", "-config", "/etc/deeplx-load-balancer-config.json"]