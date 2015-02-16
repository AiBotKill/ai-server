#!/bin/bash

docker build -t ai-server .
docker stop ai-server
docker rm ai-server
docker run -d --name ai-server --link gnatsd:gnatsd ai-server
