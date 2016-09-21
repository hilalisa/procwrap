#!/bin/bash
# $1 = source root,  $2 = output root, $3 = docker build image, $4 = output label, $5 = go package, $6 exe name

mkdir -p $2/$4
docker pull $3
docker run -v $1:/usr/local/go/src/$5 --name $4_temp $3 go install $5
docker cp $4_temp:/usr/local/go/bin/$6 $2/$4/$6-$4
docker rm $4_temp
