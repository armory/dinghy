#!/bin/bash

DBENABLED=$(yq r /opt/spinnaker/config/dinghy.yml sql.enabled)
BASEURL=$(yq r /opt/spinnaker/config/dinghy.yml sql.baseUrl)
DBNAME=$(yq r /opt/spinnaker/config/dinghy.yml sql.databaseName)
DBUSER=$(yq r /opt/spinnaker/config/dinghy.yml sql.user)
DBPASSWORD=$(yq r /opt/spinnaker/config/dinghy.yml sql.password)


if [ "$DBENABLED" != "true" ]
then
  echo "SQL is not enabled so Liquibase will not be executed"
  exit 0
fi

if [ -z "$BASEURL" ]
then
  echo "Value sql.baseUrl cannot be empty"
  exit 1
fi

if [ -z "$DBNAME" ]
then
  echo "Value sql.databaseName cannot be empty"
  exit 1
fi

if [ -z "$DBUSER" ]
then
  echo "Value sql.user cannot be empty"
  exit 1
fi

if [ -z "$DBPASSWORD" ]
then
  echo "Value sql.password cannot be empty"
  exit 1
fi

liquibase \
--classpath="/liquibase/lib/mysql.jar" \
--driver=com.mysql.cj.jdbc.Driver \
--changeLogFile="/liquibase/dbchangelog.xml" \
--url="jdbc:mysql://$BASEURL/$DBNAME" \
--username=$DBUSER \
--password=$DBPASSWORD update

