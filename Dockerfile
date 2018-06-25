# VERSION               0.0.8
# DESCRIPTION:    Metricly Heapster Docker Container
# MAINTAINER Metricly <repos@metricly.com>

FROM golang:1.8-alpine3.6

RUN apk --no-cache add --virtual .build-dependencies git make \
  && mkdir -p /go/src/github.com/metricly \
  && cd /go/src/github.com/metricly \
  && git clone https://github.com/metricly/go-client.git \
  && mkdir -p /go/src/k8s.io 

COPY . /go/src/k8s.io/heapster/

RUN cd /go/src/k8s.io/heapster \
  && make \
  && mv /go/src/k8s.io/heapster/heapster / \
  && rm -rf /go/src/k8s.io/heapster \
  && rm -rf /go/src/github.com/metricly \
  && apk del .build-dependencies

ENTRYPOINT ["/heapster"]