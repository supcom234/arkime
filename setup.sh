#!/bin/bash

# Arkime ES setup.
# set -x

echo "Checking if ElasticSearch is ready"
until curl -sS "${elastic_url}/_cluster/health?wait_for_status=yellow" > /dev/null 2>&1
do
  echo "Waiting 5s for ElasticSearch server to start"
  sleep 5
done

echo "ElasticSearch is ready! Checking Arkime index and version"

# See if database is configured already
DB_VER=`/opt/arkime/db/db.pl ${elastic_url} info | grep 'DB Version' | awk '{print $3}'`;

if [ -n "$DB_VER" ]; then
  echo "Database version is $DB_VER, skipping initial setup"
  echo "Checking if database upgrade is necessary"

  # Check the db.pl script to see what version it is
  LATEST_VER=`grep 'my $VERSION' /opt/arkime/db/db.pl | awk '{gsub(";",""); print $4}'`

  # If our script is newer, then run a database upgrade
  if [ $LATEST_VER -gt $DB_VER ]; then
    echo "Script version is $LATEST_VER which is newer, upgrading database"
    echo UPGRADE | /opt/arkime/db/db.pl ${elastic_url} upgrade
  else
    echo "Script version is $LATEST_VER which is not newer than ES index. Skipping upgrade"
    exit 0
  fi

else
  # Initial setup.
  echo "Arkime ES Database not installed, running initial setup"
  /opt/arkime/db/db.pl ${elastic_url} init

  echo "Adding admin user"
  /opt/arkime/bin/arkime_add_user.sh -c /opt/arkime/etc/config.ini ${username} "Arkime Admin" ${password} --admin
fi
