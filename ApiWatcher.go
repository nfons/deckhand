package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/apis/core"
	"os"
	"path/filepath"
)

func WatchList(resource string, resourceType runtime.Object) cache.Controller {
	var restInterface rest.Interface
	switch resourceType.(type) {
	default:
		restInterface = clientset.AppsV1().RESTClient()
	case *v1core.Service, *v1core.Secret, *v1core.ConfigMap:
		restInterface = clientset.CoreV1().RESTClient()

	}
	watchlist := cache.NewListWatchFromClient(restInterface, resource, core.NamespaceAll, fields.Everything())
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

	// Only get Secrets, Config Maps, Services only if store_all is set
	if deck_config.STORE_ALL == true {
		go WatchList(string(v1core.ResourceServices), &v1core.Service{}).Run(wait.NeverStop)
		go WatchList("secrets", &v1core.Secret{}).Run(wait.NeverStop)
		go WatchList(string(v1core.ResourceConfigMaps), &v1core.ConfigMap{}).Run(wait.NeverStop)
	}

	// IDK if I need all these
	go controller.Run(wait.NeverStop)
	go controllerSS.Run(wait.NeverStop)
	go controllerDS.Run(wait.NeverStop)

	// Watch for namespaces are different
	namespaceList := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "namespaces", core.NamespaceAll, fields.Everything())
	_, namespaceController := cache.NewInformer(
		namespaceList,
		&v1core.Namespace{},
		-1,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    namespaceAdded,
			DeleteFunc: namespaceDeleted,
		})
	namespaceController.Run(wait.NeverStop)
}

// Only used in the resource deleted field
func getResourceInfo(obj interface{}) (string, string, string) {
	switch val := obj.(type) {
	default:
		log.Panic("unknown type in deletion")
	case *v1.DaemonSet:
		return val.Name, "DaemonSet", val.Namespace
	case *v1.StatefulSet:
		return val.Name, "StatefulSet", val.Namespace
	case *v1.Deployment:
		return val.Name, "Deployment", val.Namespace
	case *v1.ReplicaSet:
		return val.Name, "ReplicaSet", val.Namespace
	case *v1core.Service:
		return val.Name, "Service", val.Namespace
	case *v1core.Secret:
		return val.Name, "Secret", val.Namespace
	case *v1core.ConfigMap:
		return val.Name, "ConfigMap", val.Namespace
	}
	return "", "", ""
}

func ResourceDeleted(obj interface{}) {
	name, rtype, namespace := getResourceInfo(obj)
	log.Println("Resource Deleted")
	filename := fmt.Sprintf("%s.%s.yaml", name, rtype)
	file := filepath.Join(createPath, namespace, filename)
	if name != "" {
		deleteFile(file)
	}
}

func ResourceAdded(obj interface{}) {
	log.Debug("Resource Added")
	SaveResource(obj)
}

func ResourceUpdated(old interface{}, obj interface{}) {
	// Because syncs also call updatefunc we will need to do this
	// create kctl deployment struct
	log.Debug("Resource Updated")
	SaveResource(obj)
}

func namespaceAdded(obj interface{}) {
	namespace := obj.(*v1core.Namespace)
	// CHECK IF EXISTS, if it doesn't then create
	dirPath := filepath.Join(createPath, namespace.Name)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// need to create this path
		log.Info("namespace: ", namespace.Name, " Added")
		if os.MkdirAll(dirPath, 0777) != nil {
			log.Error("Error creating namespace path")
		}
	}
}

func namespaceDeleted(obj interface{}) {
	namespace := obj.(*v1core.Namespace)
	log.Info("namespace", namespace.Name, " Deleted")
	// if Path doesnt exist
	dirPath := filepath.Join(createPath, namespace.Name)
	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		// need to create this path
		if os.RemoveAll(dirPath+"/") != nil {
			log.Error("Could not clear old dir struct")
		}
	} else {
		log.Error("Tried to delete a namespace that was not found in repo")
	}
}
