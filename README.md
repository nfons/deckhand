# Deck Hand
[![Go Report Card](https://goreportcard.com/badge/github.com/nfons/deckhand)](https://goreportcard.com/report/github.com/nfons/deckhand)

Kubernetes State saver for GitOps like functionality.


# What is it?
Deck Hand is an application that will Save mutable kubernetes resources such as Deployments, StatefulSets, and DaemonSets.
The objective is to save the current k8s mutable state to re-create a cluster in case of disaster.


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
Coming soon

