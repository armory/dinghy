#!/bin/bash

DBENABLED=$(yq merge /opt/spinnaker/config/dinghy* | yq r - sql.enabled)
BASEURL=$(yq merge /opt/spinnaker/config/dinghy* | yq r - sql.baseUrl)
DBNAME=$(yq merge /opt/spinnaker/config/dinghy* | yq r - sql.databaseName)
DBUSER=$(yq merge /opt/spinnaker/config/dinghy* | yq r - sql.user)
DBPASSWORD=$(yq merge /opt/spinnaker/config/dinghy* | yq r - sql.password)


if [ "$DBENABLED" != "true" ]
then
  echo "SQL is not enabled so Liquibase will not be executed"
else
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
  --driver=com.mysql.cj.jdbc.Driver \
  --changeLogFile="/liquibase/dbchangelog.xml" \
  --url="jdbc:mysql://$BASEURL/$DBNAME" \
  --username=$DBUSER \
  --password=$DBPASSWORD update

  STATUS=$?

  if [ $STATUS != 0 ]
  then
    echo "Liquibase upgrade failed, please contact armory support team."
    exit 1
  fi
fi

/opt/armory/bin/dinghy