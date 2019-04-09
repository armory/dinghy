FROM golang:1.12.1-alpine3.9 as builder

# vendor flags conflict with `go get`
# so we fetch golint before running make
# and setting the env variable
RUN apk update && apk add git make bash build-base gcc bc
RUN go get -u golang.org/x/lint/golint

ENV GO111MODULE=on GOFLAGS='-mod=vendor' GOOS=linux GOARCH=amd64
WORKDIR /opt/armory/build/
ADD ./ /opt/armory/build/
RUN make

FROM alpine:3.9

EXPOSE 8081
WORKDIR /opt/armory/bin/
RUN apk update                        \
	&& apk add ca-certificates bash   \
	&& adduser -D spinnaker           \
	&& addgroup spinnaker spinnaker
COPY --from=builder /opt/armory/build/build/dinghy /opt/armory/bin/dinghy
USER spinnaker
CMD ["/opt/armory/bin/dinghy"]
