FROM alpine

RUN apk add --update curl jq && \
    rm -rf /var/cache/apk/*

# get the latest version from github API

RUN curl -s https://api.github.com/repos/jaxxstorm/hookpick/releases/latest | jq -r '.assets[]| select(.browser_download_url | contains("linux")) | .browser_download_url' | xargs curl -L -o /tmp/hookpick.tar.gz

RUN tar zxvf /tmp/hookpick.tar.gz

RUN mv hookpick /usr/local/bin/hookpick

ENTRYPOINT ["/usr/local/bin/hookpick"] 
