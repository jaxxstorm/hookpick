FROM golang:alpine as builder
WORKDIR /go/src/github.com/jaxxstorm/hookpick
COPY . .
RUN apk add --no-cache upx ca-certificates
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-d -s -w" -o hookpick-linux-amd64 \
    && upx hookpick-linux-amd64

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
USER 1001
COPY --chown=1001 --from=builder /go/src/github.com/jaxxstorm/hookpick/hookpick-linux-amd64 .
ENTRYPOINT ["./hookpick-linux-amd64"]
CMD ["unseal"]