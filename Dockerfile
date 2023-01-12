FROM golang:1.14.15-alpine3.13 as builder

# vendor flags conflict with `go get`
# so we fetch golint before running make
# and setting the env variable
RUN apk update && apk add git make bash build-base gcc bc

ENV GO111MODULE=on GOFLAGS='-mod=vendor' GOOS=linux GOARCH=amd64
WORKDIR /opt/armory/build/
ADD ./ /opt/armory/build/

RUN make

FROM alpine:3.11

ENV PATH $PATH:/liquibase

#Liquibase download
ENV LIQUIBASE_VERSION 3.10.3

RUN apk update                        \
	&& apk add ca-certificates bash openjdk11 \
	&& adduser -D spinnaker           \
	&& addgroup spinnaker spinnaker

RUN mkdir -p /liquibase
RUN wget -O /liquibase/liquibase.tar.gz https://github.com/liquibase/liquibase/releases/download/v${LIQUIBASE_VERSION}/liquibase-${LIQUIBASE_VERSION}.tar.gz
RUN tar -xzvf /liquibase/liquibase.tar.gz -C /liquibase

# Mysql driver download
# No key published to Maven Central, using SHA256SUM
ARG MYSQL_SHA256=f93c6d717fff1bdc8941f0feba66ac13692e58dc382ca4b543cabbdb150d8bf7

RUN wget --no-verbose -O /liquibase/lib/mysql.jar https://repo1.maven.org/maven2/mysql/mysql-connector-java/8.0.19/mysql-connector-java-8.0.19.jar
RUN echo "$MYSQL_SHA256  /liquibase/lib/mysql.jar" | sha256sum -c -

# CVE-2017-18640
ARG SNAKE_YAML_SHA256=d87d607e500885356c03c1cae61e8c2e05d697df8787d5aba13484c2eb76a844
RUN rm /liquibase/lib/snakeyaml-1.24.jar
RUN wget --no-verbose -O /liquibase/lib/snakeyaml-1.26.jar https://repo1.maven.org/maven2/org/yaml/snakeyaml/1.26/snakeyaml-1.26.jar
RUN echo "$SNAKE_YAML_SHA256  /liquibase/lib/snakeyaml-1.26.jar" | sha256sum -c -

# Download changelog and copy shell script
WORKDIR /liquibase/
RUN wget https://raw.githubusercontent.com/armory/dinghy/master/liquibase/dbchangelog.xml
COPY liquibase-upgrade.sh .
RUN chmod 0755 liquibase-upgrade.sh

EXPOSE 8081
WORKDIR /opt/armory/bin/

COPY --from=builder /opt/armory/build/build/dinghy /opt/armory/bin/dinghy
USER spinnaker

CMD ["/opt/armory/bin/dinghy"]
