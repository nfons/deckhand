package main

import (
	"fmt"
	"k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/apis/core"
	"log"
	"path/filepath"
	"reflect"
)

func WatchList(resource string, resourceType runtime.Object) cache.Controller {
	watchlist := cache.NewListWatchFromClient(clientset.AppsV1().RESTClient(), resource, core.NamespaceAll, fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		resourceType,
		-1,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    ResourceAdded,
			UpdateFunc: ResourceUpdated,
			DeleteFunc: ResourceDeleted,
		})

	return controller
}

/*
	This Code Will listen to the Kube API and run file saver based on the resource returned by API Server
*/
func WatchApis() {
	controller := WatchList("deployments", &v1.Deployment{})
	controllerSS := WatchList("statefulsets", &v1.StatefulSet{})
	controllerDS := WatchList("daemonsets", &v1.DaemonSet{})

	// Only use Replica Sets if we need to since deploys == rs

	if deck_config.UseReplicaSets == true {
		go WatchList("relicasets", &v1.ReplicaSet{}).Run(wait.NeverStop)
	}

	// Only get Secrets, Config Maps, Services, Ingress only if its added
	if deck_config.STORE_ALL == true {
		//  TODO
	}

	// IDK if I need all these
	go controller.Run(wait.NeverStop)
	go controllerSS.Run(wait.NeverStop)
	go controllerDS.Run(wait.NeverStop)

}

// Only used in the resource deleted field
func getResourceInfo(obj interface{}) (string, string, string) {
	switch val := obj.(type) {
	default:
		log.Panic("unknown type in deletion")
		log.Fatal(val)
	case *v1.DaemonSet:
		return val.Name, "daemonset", val.Namespace
	case *v1.StatefulSet:
		return val.Name, "statefulset", val.Namespace
	case *v1.Deployment:
		return val.Name, "deployment", val.Namespace
	case *v1.ReplicaSet:
		return val.Name, "replicaset", val.Namespace
	}
	return "", "", ""
}

func ResourceDeleted(obj interface{}) {
	name, rtype, namespace := getResourceInfo(obj)
	filename := fmt.Sprintf("%s.%s.yaml", name, rtype)
	file := filepath.Join(createPath, namespace, filename)
	if name != "" {
		deleteFile(file)
	}
}

func ResourceAdded(obj interface{}) {
	switch val := obj.(type) {
	default:
		log.Panic("Unknown Type: ")
		log.Println(val)
		return
	case *v1.Deployment:
		log.Println("Deployment Added " + val.Name)
		namespacePath := filepath.Join(createPath, val.Namespace)
		SaveDeploy(*val, namespacePath)
	case *v1.ReplicaSet:
		if deck_config.UseReplicaSets == true {
			namespacePath := filepath.Join(createPath, val.Namespace)
			SaveRS(*val, namespacePath)
		}
	case *v1.DaemonSet:
		log.Println("DaemonSet Added " + val.Name)
		namespacePath := filepath.Join(createPath, val.Namespace)
		SaveDS(*val, namespacePath)
	case *v1.StatefulSet:
		log.Println("Satefulset Added " + val.Name)
		namespacePath := filepath.Join(createPath, val.Namespace)
		SaveSS(*val, namespacePath)
	}
}

func ResourceUpdated(old interface{}, obj interface{}) {
	// Because syncs also call updatefunc we will need to do this
	// create kctl deployment struct

	switch val := obj.(type) {
	default:
		log.Panic("Unknown Type: ")
		log.Println(val)
		return
	case *v1.Deployment:
		if reflect.DeepEqual(old, obj) == false {
			log.Println("Deployment Updated " + val.Name)
			namespacePath := filepath.Join(createPath, val.Namespace)
			SaveDeploy(*val, namespacePath)
		}
	case *v1.ReplicaSet:
		if deck_config.UseReplicaSets == true {
			if reflect.DeepEqual(old, obj) == false {
				namespacePath := filepath.Join(createPath, val.Namespace)
				SaveRS(*val, namespacePath)
			}
		}
	case *v1.DaemonSet:
		if reflect.DeepEqual(obj, old) == false {
			log.Println("Daemonset Updated " + val.Name)
			namespacePath := filepath.Join(createPath, val.Namespace)
			SaveDS(*val, namespacePath)
		}
	case *v1.StatefulSet:
		if reflect.DeepEqual(old, obj) == false {
			log.Println("Statefulset Updated " + val.Name)
			namespacePath := filepath.Join(createPath, val.Namespace)
			SaveSS(*val, namespacePath)
		}
	}
}
