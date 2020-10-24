#!/bin/bash

docker build --tag eoip .
docker run --name eoip -it eoip ls -al
docker cp eoip:/root/eoip/eoip ./
#docker cp eoip:/root/eoip/evlan ./
docker rm eoip
docker rmi eoip
