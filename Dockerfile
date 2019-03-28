FROM golang:1.11-alpine AS builder

RUN apk --update add make gcc git dep curl

WORKDIR /go/src/github.com/influxdata/

#COPY . telegraf
#RUN cd telegraf && make

#FROM alpine:3.9

#RUN apk add --no-cache net-snmp-tools procps lm_sensors dumb-init iputils

#RUN addgroup -S telegraf && adduser -S -s /bin/false -G telegraf telegraf

#WORKDIR /telegraf
#COPY --chown=telegraf:telegraf --from=builder /go/src/github.com/influxdata/telegraf/telegraf /telegraf
