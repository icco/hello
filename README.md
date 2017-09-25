# hello

A simple hello.

## Build

```
docker build -t gcr.io/icco-cloud/hello:$(date +%Y.%m.%d) .
```

## Deploy

```
gcloud docker -- push gcr.io/icco-cloud/hello:$(date +%Y.%m.%d)
```
