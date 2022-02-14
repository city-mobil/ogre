FROM golang:rc-alpine3.15 as builder

ARG CI_COMMIT_SHORT_SHA=""
ARG CI_COMMIT_TAG=""

ENV GIT_TERMINAL_PROMPT=1

RUN apk --update add git less openssh && \
    rm -rf /var/lib/apt/lists/* && \
    rm /var/cache/apk/*

WORKDIR /go/src/city-mobil/ogre
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/ogre -ldflags="-X 'main.buildDateTime=$(date +%Y-%m-%dT%H:%M:%S%z)' -X 'main.gitCommit=$CI_COMMIT_SHORT_SHA' -X 'main.versionTag=$CI_COMMIT_TAG'  -X 'main.buildAuthor=$GITLAB_USER_LOGIN'" cmd/ogre/ogre.go

FROM alpine:3.15

WORKDIR /usr/bin
COPY --from=builder /go/src/city-mobil/ogre/bin/ogre ./
RUN adduser --no-create-home --disabled-password --gecos "" ogre
RUN chown -R ogre:ogre ./ogre
RUN chown -R ogre:ogre /var/log/

USER ogre
RUN chmod +x ./ogre

EXPOSE 8888

ENTRYPOINT ["/usr/bin/ogre"]
