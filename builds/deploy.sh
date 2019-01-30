#!/usr/bin/env bash

if [ $CIRCLE_BRANCH == 'master' ]; then
     docker tag deckhand quay.io/nfons/deckhand:latest
     docker push quay.io/nfons/deckhand:latest
else
     docker tag deckhand quay.io/nfons/deckhand:$CIRCLE_BRANCH
     docker push quay.io/nfons/deckhand:$CIRCLE_BRANCH
fi