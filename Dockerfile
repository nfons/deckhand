FROM golang:stretch

ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

WORKDIR /go/src/github.com/nfons/deckhand
ADD . .

RUN dep ensure
RUN go build -o deckhand

ENTRYPOINT ["sh", "-c", "ssh"]
