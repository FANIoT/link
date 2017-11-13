# ISRC Push.Go
## Introduction
Push service of ISRC platform

## Running
MongoDB

```sh
docker run -ti --rm -p 27017:27017 mongo
```

Vernemq

```sh
docker run -e "DOCKER_VERNEMQ_ALLOW_ANONYMOUS=on" --rm -ti -p 1883:1883 --name vernemq1 erlio/docker-vernemq
```
