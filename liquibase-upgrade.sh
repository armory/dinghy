#!/bin/bash

BASEURL=$1
DBNAME=$2
DBUSER=$3
DBPASSWORD=$4

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

exit 0