#### BUILDER

FROM golang:stretch AS builder

WORKDIR /go/src/github.com/nfons/deckhand
ADD . .

# not sure if we need this
RUN rm -rf .git/

RUN go get -d -v ./...

RUN go build -o /go/bin/deckhand


## Actual small img
FROM alpine:3.6

WORKDIR /go/bin

COPY --from=builder /go/bin/deckhand /go/bin/deckhand


ENTRYPOINT ["sh", "-c", "ssh"]
