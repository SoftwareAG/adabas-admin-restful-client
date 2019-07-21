#!/bin/bash

echo "Start admin client $*"

ADABAS_ADMIN_HOME=`pwd`
SRCDIR=${ADABAS_ADMIN_HOME}/src
ENABLE_DEBUG=1
LOGPATH=`pwd`/logs
if [ ! -r $LOGPATH ]; then
   mkdir $LOGPATH
fi
export ENABLE_DEBUG LOGPATH ADABAS_ADMIN_HOME
go run cmd/adabas-rest-client/main.go $*
