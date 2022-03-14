# How to build and deploy

## Build

google cloud run build
https://cloud.google.com/run/docs/building/containers#docker


### coaches (caddy)

```
docker build -t us-west1-docker.pkg.dev/teams-view/container/coaches:latest . --progress plain --no-cache
```

## Deploy

google cloud run deploy
https://cloud.google.com/run/docs/deploying#revision

## .gcloud folder

Contents for managing build and run from `gcloud` cli

## .coaches folder

Contents for generating `X-Coach` header token
