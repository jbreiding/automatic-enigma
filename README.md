# How to build and deploy

## Build

google cloud run build
https://cloud.google.com/run/docs/building/containers#docker

### teams-view (main.go)
http://us-west1-docker.pkg.dev/teams-view/container/teams-view

```
docker build -t us-west1-docker.pkg.dev/teams-view/container/teams-view:latest . --progress plain --no-cache
```

### coaches (caddy)
http://us-west1-docker.pkg.dev/teams-view/container/coaches

```
docker build -t us-west1-docker.pkg.dev/teams-view/container/coaches:latest . --progress plain --no-cache
```

## Deploy

google cloud run deploy
https://cloud.google.com/run/docs/deploying#revision