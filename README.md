[![Go](https://github.com/middlewaregruppen/container-demo/actions/workflows/go.yaml/badge.svg)](https://github.com/middlewaregruppen/container-demo/actions/workflows/go.yaml)

# container-demo
Container-demo is an application that you can use to show-case various container capabilities such as resource limits, readiness probes and logging.

![](img/screenshot.png)

## Run on Kubernetes
```
kubectl apply -f https://raw.githubusercontent.com/middlewaregruppen/container-demo/main/deploy.yaml
```

## Run on Docker
```
docker run -p 8080:8080 ghcr.io/middlewaregruppen/container-demo:latest
```