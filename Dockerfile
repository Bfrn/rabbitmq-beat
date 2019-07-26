FROM golang

LABEL maintainer="Björn Franke"

RUN mkdir -p ${GOPATH}/src/github.com/elastic && \
    git clone https://github.com/elastic/beats ${GOPATH}/src/github.com/elastic/beats && \
    apt-get update && \
    apt-get install -y virtualenv

WORKDIR $GOPATH/src/hummer/rabbitmq-beat

COPY ./* ./

RUN make update

CMD rabbitmq-beat -c rabbitmq-beat.yml -e -d "*"