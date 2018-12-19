FROM golang:stretch

ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

WORKDIR /go/src/github.com/nfons/deckhand
ADD . .

# not sure if we need this
RUN rm -rf .git/

RUN dep ensure
RUN go build -o deckhand

ENTRYPOINT ["sh", "-c", "ssh"]
