#!/usr/bin/env bash
set -e

if [ "$1" == "" ]
then
  echo "error: have to give an app name to build like './docker.sh permissions|play'"
  exit -1
fi

if [ "$2" == "" ]
then
  echo "error: you need to specify image version. like latest|0.0.1"
  exit -1
fi

if [ "$3" == "" ]
then
  echo "error: you need to specify environment name. like aws|otokoc|otokoc_dev"
  exit -1
fi

APP_NAME=$(echo $1 | sed 's/\([a-z0-9]\)\([A-Z]\)/\1-\2/g' | tr '[:upper:]' '[:lower:]'})
IMAGE_NAME="iugo/${APP_NAME}"
IMAGE_VERSION='0.0.1'

if [ "$2" != "" ]
then
  IMAGE_VERSION=$2
fi

if [ "$3" == "aws" ]
then
  docker context use aws_mysql
  DOCKER_REGISTRY="172.31.3.119:5000" # iugo structure registry
fi
if [ "$3" == "iua" ]
then
  DOCKER_REGISTRY="172.31.4.253:5000"
  docker context use aws_iua_1
fi

if [ "$3" == "otokoc_dev" ]
then
  docker context use otokoc_dev_2
  DOCKER_REGISTRY="10.180.32.21:5000" # otokoc dev registry
fi

if [ "$3" == "getir_order" ]
then
  docker context use aws_gtr_1
  DOCKER_REGISTRY="172.31.15.237:5000" # getir prod registry
fi

if [ "$3" == "otokoc_prod" ]
then
  docker context use otokoc_drv_6
  DOCKER_REGISTRY="10.180.43.16:5000" # otokoc prod registry
fi

if [ "$3" == "iugo_dev" ]
then
  docker context use iugo_dev
  DOCKER_REGISTRY="172.31.3.119:5000" # iugo dev registry
fi

if [ "$3" == "local" ]
then
  docker context use default
  DOCKER_REGISTRY="127.0.0.1:5000" # iugo dev registry
fi

IMAGE_TAG="$DOCKER_REGISTRY/$IMAGE_NAME:$IMAGE_VERSION"

# docker build -t ${IMAGE_TAG} -f 'Dockerfile_kafka' ./ --build-arg app_name=${1}

# docker build --no-cache -t ${IMAGE_TAG} ./ --build-arg app_name=${1}

docker build -t ${IMAGE_TAG} ./

docker push ${IMAGE_TAG}

if [ "$4" != "" ] && [ "$4" != "skip_service_update" ]
then
  docker service update $4 --image $IMAGE_TAG
fi
