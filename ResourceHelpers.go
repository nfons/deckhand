package main

import (
	"fmt"
	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	"path/filepath"
	"reflect"
)

// We need to redo this logic to get rid of some of this damn copy and paste crap...ideally I would like to
// type cast all these to some struct so I can get the relevant fields from it
func SaveResource(obj interface{}) {
	resource := KubeStruct{}
	var namespacePath string
	switch val := obj.(type) {
	default:
		log.Println("Unknown type in SaveResource : " + reflect.TypeOf(val).String())
		return
	case *v1.DaemonSet:
		resource.ApiVersion = "apps/v1"
		resource.Kind = "DaemonSet"
		resource.Metadata = val.ObjectMeta
		resource.Spec = val.Spec
		namespacePath = filepath.Join(createPath, val.Namespace)
	case *v1.StatefulSet:
		resource.ApiVersion = "apps/v1"
		resource.Kind = "StatefulSet"
		resource.Metadata = val.ObjectMeta
		resource.Spec = val.Spec
		namespacePath = filepath.Join(createPath, val.Namespace)
	case *v1.Deployment:
		resource.ApiVersion = "apps/v1"
		resource.Kind = "Deployment"
		resource.Metadata = val.ObjectMeta
		resource.Spec = val.Spec
		namespacePath = filepath.Join(createPath, val.Namespace)
	case *v1.ReplicaSet:
		resource.ApiVersion = "apps/v1"
		resource.Kind = "ReplicaSet"
		resource.Metadata = val.ObjectMeta
		resource.Spec = val.Spec
		namespacePath = filepath.Join(createPath, val.Namespace)
	case *v1core.Service:
		resource.ApiVersion = "v1"
		resource.Kind = "Service"
		resource.Metadata = val.ObjectMeta
		resource.Spec = val.Spec
		namespacePath = filepath.Join(createPath, val.Namespace)
	case *v1core.Secret:
		resource.ApiVersion = "v1"
		resource.Kind = val.Kind
		resource.Metadata = val.ObjectMeta
		resource.Kind = "Secret"
		resource.Data = val.Data
		namespacePath = filepath.Join(createPath, val.Namespace)
	case *v1core.ConfigMap:
		resource.ApiVersion = "v1"
		resource.Metadata = val.ObjectMeta
		resource.Kind = "ConfigMap"
		resource.Data = val.Data
		namespacePath = filepath.Join(createPath, val.Namespace)
	}
	yamlName := fmt.Sprintf("%s/%s.%s.yaml", namespacePath, resource.Metadata.Name, resource.Kind)

	// save the file
	marshalYaml, err := yaml.Marshal(resource)
	if err != nil {
		log.Panic(err)
	}
	log.Debug("Saving:", resource.Kind, resource.Metadata.Name)
	saveFile(marshalYaml, yamlName)

}
