FROM golang:1.14-alpine as build
RUN apk add --no-cache make gcc musl-dev linux-headers git
ENV GO111MODULE=on

WORKDIR apps/pharos
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o /usr/local/bin/iotex-pharos .

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=build /usr/local/bin/iotex-pharos /usr/local/bin
CMD [ "iotex-pharos"]
