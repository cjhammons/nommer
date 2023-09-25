#!/bin/bash
REV="latest"
if [ "$1" != "" ]; then
  REV=$1
fi
echo "Using tag $REV"

docker build --platform linux/amd64 -t cjhammons/nommer:$REV .
docker push cjhammons/nommer:$REV #requires auth
