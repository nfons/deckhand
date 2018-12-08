package main

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KubeStruct struct {
	ApiVersion string               `json:"apiVersion"`
	Kind       string               `json:"kind"`
	Metadata   v1.ObjectMeta        `json:"metadata,omitempty"`
	Spec       interface{} `json:"spec,omitempty"`
}
