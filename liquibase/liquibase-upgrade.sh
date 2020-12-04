#!/bin/bash

DBENABLED=$(yq r /opt/spinnaker/config/dinghy.yml sql.enabled)
BASEURL=$(yq r /opt/spinnaker/config/dinghy.yml sql.baseUrl)
DBNAME=$(yq r /opt/spinnaker/config/dinghy.yml sql.databaseName)
DBUSER=$(yq r /opt/spinnaker/config/dinghy.yml sql.user)
DBPASSWORD=$(yq r /opt/spinnaker/config/dinghy.yml sql.password)

#I need to add validations for empty values since all are mandatory

liquibase \
--classpath="/liquibase/lib/mysql.jar" \
--driver=com.mysql.cj.jdbc.Driver \
--changeLogFile="/liquibase/dbchangelog.xml" \
--url="jdbc:mysql://$BASEURL/$DBNAME" \
--username=$DBUSER \
--password=$DBPASSWORD update
