#!/bin/sh -x
date=$(/bin/date +"%Y-%m-%d")
IMAGE=android_docker

# TODO(anyone): Delete this crap and hardcode -f once you upgrade docker.
/usr/bin/docker --version \
  | /bin/grep 'Docker version 1.12.2, build bb80604' \
    && TAG_FLAG="" \
    || TAG_FLAG="-f"

if [ "$1" = "-n" ]; then
  cache="--no-cache=true"
  shift
else
  cache="--no-cache=false"
fi


/usr/bin/docker build $cache -t ${IMAGE}:$date .
if [ $? = 0 ]; then
  /usr/bin/docker tag $TAG_FLAG ${IMAGE}:$date ${IMAGE}:latest
  exit 0
fi
exit 1
