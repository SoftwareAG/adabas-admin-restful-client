#!/bin/bash


ADABAS_ADMIN_HOME=`pwd`
SRCDIR=${ADABAS_ADMIN_HOME}/src
cd ${ADABAS_ADMIN_HOME}/src/softwareag.com
ENABLE_DEBUG=1
GOPATH=$GOPATH:$ADABAS_ADMIN_HOME
LOGPATH=`pwd`/logs
if [ ! -r $LOGPATH ]; then
   mkdir $LOGPATH
fi
export ENABLE_DEBUG LOGPATH ADABAS_ADMIN_HOME
GOPATH=$GOPATH go run cmd/client.go $*
