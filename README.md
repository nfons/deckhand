# Deck Hand
[![Build Status](https://travis-ci.org/nfons/deckhand.svg?branch=master)](https://travis-ci.org/nfons/deckhand)
[![Go Report Card](https://goreportcard.com/badge/github.com/nfons/deckhand)](https://goreportcard.com/report/github.com/nfons/deckhand)
[![License](https://img.shields.io/github/license/nfons/deckhand.svg)](https://github.com/nfons/deckhand/blob/master/LICENSE)
[![Release](https://img.shields.io/github/release-pre/nfons/deckhand.svg)](https://github.com/nfons/deckhand/releases)
[![GolangCI](https://golangci.com/badges/github.com/nfons/deckhand.svg)](https://golangci.com/badges/github.com/nfons/deckhand)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fgo-swagger%2Fgo-swagger.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fgo-swagger%2Fgo-swagger?ref=badge_shield)
[![Docker Repository on Quay](https://quay.io/repository/nfons/deckhand/status "Docker Repository on Quay")](https://quay.io/repository/nfons/deckhand)

## What is it?
Deck Hand is an application that will save mutable kubernetes resources states such as Deployments, StatefulSets, and DaemonSets into kubectl compliant files to a git repo.
The objective is to save the current k8s mutable state to re-create a cluster in case of disaster, replication or any other need.

![General Architecture](https://i.imgur.com/jNPSMhE.png)

Then @ later time

![Gen arch1](https://i.imgur.com/hNyZ4NF.png)

## ENV VARS
The Application can be configured using env vars. Each ENV var is prefixed with DECK_*
Current ENV Vars:


| ENV Var Name  | Type  |  Default | Required  |  Comment |
|---|---|---|---|---|
|  DECK_GIT_REPO |  string | nil   | ✔  | Git Repo you want to save states to   |
|  DECK_SYNCINTERVAL | string   | 30s   | ❌  | Must be valid go time parse duration format  https://golang.org/pkg/time/#ParseDuration |
|  DECK_CLUSTER_NAME | string  | dev  | ❌  |  cluster name you want to save under  |
| DECK_USE_REPLICA_SETS| bool | F | ❌ |  If you want to save replica sets as well (not recommended) ||
|DECK_SSH_KEY | string | nil | ✔ (sort of) | SSH Private key you want to use to connect to git repo |
|DECK_GIT_USER| string| nil | ❌ | Git username you will use if using https git|
|DECK_GIT_PASSWORD|string|nil|✔ (sort of) | Git password if using https git|


# Getting Started

We will assume you have a cluster running (like [minikube](https://kubernetes.io/docs/setup/minikube/) , or [docker-k8s](https://docs.docker.com/docker-for-mac/kubernetes/))

#### A) If using SSH (recommended) for git:
1. Create and associate a SSH key for your git repo by following this guide [HERE](https://help.github.com/articles/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent/)
2. Next Add that key  as a k8s secret:

    `kubectl create secret git-ssh-key --from-file=[location if your ssh key private key]`
    
3. Edit the following `builds/deployments.yaml` to fit your needs:
    
            apiVersion: apps/v1
            kind: Deployment
            metadata:
              name: deckhand
              labels:
                app: deckhand
            spec:
              replicas: 1
              selector:
                matchLabels:
                  app: deckhand
              template:
                metadata:
                  labels:
                    app: deckhand
                spec:
                  containers:
                    - name: deckhand
                      image: nfons/deckhand:latest
                      ports:
                        - containerPort: 8080
                      env:
                        - name: DECK_GIT_REPO
                          value: [[ YOUR SSH GIT REPO... i.e git@github.com:nfons/deckhand-example.git ]]
                        - name: DECK_SSH_KEY
                          valueFrom:
                            secretKeyRef:
                              name: git-ssh-key
                              key: [[ NAME OF YOUR SSH KEY FILE ]]
    
Note: Ensure Your `DECK_GIT_REPO` is the ssh format (i.e git@(yourhost)

#### B) If using HTTPS:
Ideally you would want to store the git username and password as secrets, but simplicity we will disregard that.

3. Edit the deployment yaml file:

            apiVersion: apps/v1
            kind: Deployment
            metadata:
              name: deckhand
              labels:
                app: deckhand
            spec:
              replicas: 1
              selector:
                matchLabels:
                  app: deckhand
              template:
                metadata:
                  labels:
                    app: deckhand
                spec:
                  containers:
                    - name: deckhand
                      image: nfons/deckhand:latest
                      ports:
                        - containerPort: 8080
                      env:
                        - name: DECK_GIT_REPO
                          value: [[ YOUR SSH GIT REPO... i.e git@github.com:nfons/deckhand-example.git ]]
                        - name: DECK_GIT_USER
                          value: [[ YOUR USER NAME ]]
                        - name: DECK_GIT_PASSWORD
                          value: [[ YOUR PASSWORD ]]
                          
---

4. `kubectl create -f builds/deployment.yaml`

5. Profit! After a short delay, you should start seeing your k8s state synced with your git repo (Take a look at the example repo: [here](https://github.com/nfons/deckhand-example) to view ops repo structure)

6. Add Some new deployments, after a short delay, you will see that deployment also synced to your git repo

# Why is K8s the source of truth instead of Git Repo?
DeckHands approach is to assume k8s as the source of truth. Traditional
GitOps (or rather most common) assumes Git repo to be the source of
truth.

![From Weave gitops](https://i.imgur.com/UAgBM0i.png)


**This approach has some draw backs for certain teams:**

- Current CI/CD pipelines need to be altered to be based on the config
  updating logic
- Need pipelines for all resources. even one-off resources like a vendor deployment

- False source of truth. If resources get updated outside of pipeline
  (i.e via developer kubectl) resource state in k8s defers from git repo

