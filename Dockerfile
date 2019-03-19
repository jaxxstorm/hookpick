FROM golang:1.10-alpine3.7 as build

ARG GOOS=linux
ARG GOARCH=amd64

RUN apk add curl git --update --no-cache
RUN curl https://glide.sh/get | sh

ADD . $GOPATH/src/github.com/jaxxstorm/hookpick

RUN cd $GOPATH/src/github.com/jaxxstorm/hookpick \
 && glide install \
 && env GOOS=${GOOS} GOARCH=${GOARCH} go build -o hookpick main.go \
 && mv ./hookpick /

FROM alpine:3.7
COPY --from=build /hookpick /usr/sbin/hookpick
VOLUME ["/root"]
ENTRYPOINT ["hookpick"]