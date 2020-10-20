FROM golang:1.13-alpine3.10 as build

RUN apk add --no-cache --update git make

RUN mkdir /build
WORKDIR /build
COPY . .
RUN make build


FROM alpine:3.10

RUN apk add --no-cache --update curl ca-certificates

USER nobody

COPY --from=build /build/bin/server /bin/server
EXPOSE 8080

ENTRYPOINT  [ "/bin/server" ]
