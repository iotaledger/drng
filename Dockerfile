FROM golang:1.14.1-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh gcc musl-dev
COPY . "/go/src/github.com/drand/drand"
WORKDIR "/go/src/github.com/drand/drand"
# from https://dev.to/plutov/docker-and-go-modules-3kkn
ENV GO111MODULE=on

COPY go.mod .
COPY go.sum .

RUN go install -ldflags "-X main.version=`git describe --tags` -X main.buildDate=`date -u +%d/%m/%Y@%H:%M:%S` -X main.gitCommit=`git rev-parse HEAD`" 

# remove sources for compactness
RUN rm -rf "/go/src/github.com/drand/drand"

WORKDIR /

ENTRYPOINT ["drand"]

