## Deck Hand
[![Go Report Card](https://goreportcard.com/badge/github.com/nfons/deckhand)](https://goreportcard.com/report/github.com/nfons/deckhand)

Kubernetes State saver for GitOps like functionality.


# What is it?
Deck Hand is an application that will Save mutable kubernetes resources such as Deployments, StatefulSets, and DaemonSets.
The objective is to save the current k8s mutable state to re-create a cluster in case of disaster.

![General Architecture](https://i.imgur.com/jNPSMhE.png)

![Gen arch1](https://i.imgur.com/hNyZ4NF.png)

## ENV VARS
The Application can be configured using env vars. Each ENV var is prefixed with DECK_*
Current ENV Vars:


| ENV Var Name  | Type  |  Default | Required  |  Comment |
|---|---|---|---|---|
|  DECK_GIT_REPO |  string | nil   | Y  | Git Repo you want to save states to   |
|  DECK_SYNCINTERVAL | string   | 30s   | N  | Must be valid go time parse duration format  https://golang.org/pkg/time/#ParseDuration |
|  DECK_CLUSTER_NAME | string  | dev  | N  |  cluster name you want to save under  |
| DECK_USE_REPLICA_SETS| bool | F | N |  If you want to save replica sets as well (not recommended) ||
|DECK_SSH_KEY | string | nil | Y| SSH Private key you want to use to connect to git repo |
|DECK_GIT_USER| string| nil | N | Git username you will use if using https git|
|DECK_GIT_PASSWORD|string|nil|N| Git password if using https git|



# Getting Started

We will assume you have a cluster running (like [minikube](https://kubernetes.io/docs/setup/minikube/) , or [docker-k8s](https://docs.docker.com/docker-for-mac/kubernetes/))

#### A) If using SSH (recommended) for git:
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

4. `kubectl create -f deployments/deployment.yaml`

5. Profit! After a short delay, you should start seeing your k8s state synced with your git repo (Take a look at the example repo: [here](https://github.com/nfons/deckhand-example) to view ops repo structure)

6. Add Some new deployments, after a short delay, you will see that deployment also synced to your git repo
