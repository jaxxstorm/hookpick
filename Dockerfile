FROM golang:1.9

WORKDIR /go/src/github.com/jaxxstorm/hookpick

COPY . .

RUN go get -v github.com/Masterminds/glide

RUN cd $GOPATH/src/github.com/Masterminds/glide && git checkout tags/v0.12.3 && go install && cd -

RUN ls .

RUN glide install

RUN go build -o hookpick main.go

ENTRYPOINT ["./hookpick"] 
