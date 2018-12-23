# Deck Hand
[![Go Report Card](https://goreportcard.com/badge/github.com/nfons/deckhand)](https://goreportcard.com/report/github.com/nfons/deckhand)

Kubernetes State saver for GitOps like functionality.


# What is it?
Deck Hand is an application that will Save mutable kubernetes resources such as Deployments, StatefulSets, and DaemonSets.
The objective is to save the current k8s mutable state to re-create a cluster in case of disaster.

![General Architecture](https://i.imgur.com/12ybhUg.png)

## ENV VARS
The Application can be configured using env vars. Each ENV var is prefixed with DECK_*
Current ENV Vars:


| ENV Var Name  | Type  |  Default | Required  |  Comment |
|---|---|---|---|---|
|  GIT_REPO |  string | nil   | Y  | Git Repo you want to save states to   |
|  SYNCINTERVAL | string   | 30s   | N  | Must be valid go time parse duration format  https://golang.org/pkg/time/#ParseDuration |
|  CLUSTER_NAME | string  | dev  | N  |  cluster name you wnat to save under  |
| USE_REPLICA_SETS| bool | F | N |  If you want to save replica sets as well (not recommended) ||
|SSH_KEY | string | nil | Y| SSH Private key you want to use to connect to git repo |


# Getting Started

We will assume you have a cluster running (like minikube, or docker-k8s)

1. Create and associate a SSH key for your git repo by following this guide [HERE](https://help.github.com/articles/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent/)
2. Next Add that key  as a k8s secret:

    `kubectl create secret git-ssh-key --from-file=[location if your ssh key private key]`
3. Edit the following `deployments/deployments.yaml` to fit your needs:

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

4. `kubectl create -f deployments/deployment.yaml`

5. Profit! After a short delay, you should start seeing your k8s state synced with your git repo


