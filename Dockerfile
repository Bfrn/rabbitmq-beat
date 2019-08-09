FROM golang AS building-stage

RUN apt-get update && \
    apt-get install -y apt-utils virtualenv build-essential python-pip && \
    mkdir -p ${GOPATH}/src/github.com/elastic && \
    git clone https://github.com/elastic/beats ${GOPATH}/src/github.com/elastic/beats && \
    go get github.com/streadway/amqp

COPY . ${GOPATH}/src/hummer/rabbitmq-beat

WORKDIR  ${GOPATH}/src/hummer/rabbitmq-beat

RUN go clean && \
    make setup && \
    sed -i 's/go build $(GOBUILD_FLAGS)/CGO_ENABLED=0 go build $(GOBUILD_FLAGS)/' ./vendor/github.com/elastic/beats/libbeat/scripts/Makefile && \
    make


FROM alpine

LABEL maintainer="Björn Franke"

COPY --from=building-stage /go/src/hummer/rabbitmq-beat/rabbitmq-beat /rabbitmq/rabbitmq-beat

COPY --from=building-stage /go/src/hummer/rabbitmq-beat/rabbitmq-beat.docker.yml /rabbitmq/rabbitmq-beat.docker.yml

WORKDIR /rabbitmq

RUN chmod go-w /rabbitmq/rabbitmq-beat.docker.yml

CMD ./rabbitmq-beat -c rabbitmq-beat.docker.yml -e -d "*"
