FROM golang:1.14.0-alpine3.11 as builder

RUN apk update && apk add git

WORKDIR /src

ADD . ./

RUN go build

# final image
FROM alpine:3.9

RUN apk update && apk add ca-certificates --no-cache

RUN mkdir -p /app
RUN mkdir -p /app/key

COPY --from=builder /src/proxy /app/proxy

EXPOSE 3000
VOLUME ["/app"]
CMD ["/app/proxy"]