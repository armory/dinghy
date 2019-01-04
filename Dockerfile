FROM alpine:3.6

EXPOSE 8081

WORKDIR /usr/local/bin/

ADD ./build/main /usr/local/bin/main
RUN apk update
RUN apk add ca-certificates
RUN apk add bash

RUN adduser -D spinnaker && \
  addgroup spinnaker spinnaker

USER spinnaker

ENTRYPOINT ["/usr/local/bin/main"]
