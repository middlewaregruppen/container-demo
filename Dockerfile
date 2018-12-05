FROM golang:alpine AS build-env
RUN  apk add --no-cache git mercurial

WORKDIR /go/src/gitlab.com/middlewaregruppen/container-demo
COPY ./ .

RUN go get -d -v  ./... && \
CGO_ENABLED=0 GOOS=linux go build -o ./bin/demo ./cmd/demo

FROM scratch

COPY --from=build-env /go/src/gitlab.com/middlewaregruppen/container-demo/bin/demo ./
COPY --from=build-env /go/src/gitlab.com/middlewaregruppen/container-demo/ui ./ui

EXPOSE 8080

CMD ["./demo"]
