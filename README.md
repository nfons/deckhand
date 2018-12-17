# Deck Hand
[![Go Report Card](https://goreportcard.com/badge/github.com/nfons/deckhand)](https://goreportcard.com/report/github.com/nfons/deckhand)

Kubernetes State saver for GitOps like functionality.


# What is it?
Deck Hand is an application that will Save mutable kubernetes resources such as Deployments, StatefulSets, and DaemonSets.
The objective is to save the current k8s mutable state to re-create a cluster in case of disaster.


## ENV VARS
The Application can be configured using env vars. Each ENV var is prefixed with DECK_*
Current ENV Vars:


    type DeckConfig struct {
        GitRepo        string `envconfig:"GIT_REPO" required:"true"`
        SyncInterval   string `default:"30s"`
        ClusterName    string `envconfig:"CLUSTER_NAME" default:"dev"`
        UseReplicaSets bool   `encconfig:"USE_REPLICA_SETS" default:"false"`
    }

# Getting Started
Coming soon

# Design Decisions
## Why only mutable resources?
   We consider things like Services, Secrets , etc to be "immutable" in the sense that the probability of these changing during a CI/CD pipeline is slim.
  Services and its like, will unlikely to change from deployment to deployment, as such it's better to store them in a static repo than one that can be dynamic.
