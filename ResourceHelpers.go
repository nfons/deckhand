package main

import (
	"fmt"
	"github.com/ghodss/yaml"
	"k8s.io/api/apps/v1"
	"log"
)

func SaveSS(deploy v1.StatefulSet, path string) {
	yamlName := fmt.Sprintf("%s/%s.statefulset.yaml", path, deploy.Name)

	// create kctl deployment struct
	deployment := KubeStruct{}

	deployment.ApiVersion = "apps/v1"
	deployment.Kind = "StatefulSet"

	deployment.Metadata = deploy.ObjectMeta
	deployment.Spec = deploy.Spec

	// save the deployment file
	marshalYaml, err := yaml.Marshal(deployment)
	if err != nil {
		log.Panic(err)
	}
	saveFile(marshalYaml, yamlName)
}

func SaveDeploy(deploy v1.Deployment, path string) {
	yamlName := fmt.Sprintf("%s/%s.deploy.yaml", path, deploy.Name)

	// create kctl deployment struct
	deployment := KubeStruct{}

	deployment.ApiVersion = "apps/v1"
	deployment.Kind = "Deployment"

	deployment.Metadata = deploy.ObjectMeta
	deployment.Spec = deploy.Spec

	// save the deployment file
	marshalYaml, err := yaml.Marshal(deployment)
	if err != nil {
		log.Panic(err)
	}
	saveFile(marshalYaml, yamlName)
}

func SaveDS(deploy v1.DaemonSet, path string) {
	yamlName := fmt.Sprintf("%s/%s.daemonset.yaml", path, deploy.Name)

	// create kctl deployment struct
	deployment := KubeStruct{}

	deployment.ApiVersion = "apps/v1"
	deployment.Kind = "DaemonSet"

	deployment.Metadata = deploy.ObjectMeta
	deployment.Spec = deploy.Spec

	// save the deployment file
	marshalYaml, err := yaml.Marshal(deployment)
	if err != nil {
		log.Panic(err)
	}
	saveFile(marshalYaml, yamlName)
}

func SaveRS(deploy v1.ReplicaSet, path string) {
	yamlName := fmt.Sprintf("%s/%s.replicaset.yaml", path, deploy.Name)

	// create kctl deployment struct
	deployment := KubeStruct{}

	deployment.ApiVersion = "apps/v1"
	deployment.Kind = "ReplicaSet"

	deployment.Metadata = deploy.ObjectMeta
	deployment.Spec = deploy.Spec

	// save the deployment file
	marshalYaml, err := yaml.Marshal(deployment)
	if err != nil {
		log.Panic(err)
	}
	saveFile(marshalYaml, yamlName)
}
