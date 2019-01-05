package main

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Kube struct is just a mock of Kubectl file struct
type KubeStruct struct {
	ApiVersion string        `json:"apiVersion"`
	Kind       string        `json:"kind"`
	Metadata   v1.ObjectMeta `json:"metadata,omitempty"`
	Spec       interface{}   `json:"spec,omitempty"`
	Type       string        `json:"type,omitempty"`
	Data       interface{}   `json:"data,omitempty"`
}
