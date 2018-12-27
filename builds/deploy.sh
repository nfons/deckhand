#!/usr/bin/env bash

if [ $TRAVIS_BRANCH == 'master' ]; then
     docker tag deckhand quay.io/nfons/deckhand:latest
     docker push quay.io/nfons/deckhand:latest
else
     docker tag deckhand quay.io/nfons/deckhand:$TRAVIS_BRANCH
     docker push quay.io/nfons/deckhand:$TRAVIS_BRANCH
fi