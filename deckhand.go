package main

import (
	"fmt"
	"github.com/ghodss/yaml" //this is better than regular yaml
	"io/ioutil"
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
)

func main() {
	//set kubeconfig, probably will disable this later
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")


	//Get Cluster Name
	cluserName := os.Getenv("DECK_CLUSTER_NAME")

	if cluserName == "" {
		cluserName = "dev"
	}

	createPath := filepath.Join("states", cluserName)

	//check if the cluster named folder exists
	if _, err := os.Stat(createPath); os.IsNotExist(err) {
		//need to create this path
		os.MkdirAll(createPath, 0777)
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset, to use for our api
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})

	//for each namespace, we will iterate and get deployments
	for _, val := range namespaces.Items{
		log.Printf(val.Name)
		Deployments, err := clientset.AppsV1().Deployments(val.Name).List(metav1.ListOptions{})
		//we should probably also get rs, ss, and ds as those would change as well.
		//we will get them, but leave them blank for now
		clientset.AppsV1().DaemonSets(val.Name).List(metav1.ListOptions{})
		clientset.AppsV1().StatefulSets(val.Name).List(metav1.ListOptions{})
		clientset.AppsV1().ReplicaSets(val.Name).List(metav1.ListOptions{})


		if err != nil {
			log.Fatal(err)
		}

		//create NS  dir if not exist
		namespacePath := filepath.Join(createPath, val.Name)

		//create the namespace path if it doesn't exist
		if _,exist := os.Stat(namespacePath); os.IsNotExist(exist) {
			os.MkdirAll(filepath.Join(namespacePath), 0777)
		}

		//save the namespace as well
		namespace := KubeStruct{}
		namespace.ApiVersion = "v1"
		namespace.Kind = "Namespace"
		namespace.Metadata = val.ObjectMeta
		namespace.Spec = val.Spec
		yamlName := fmt.Sprintf("%s/NAMESPACE.yaml", namespacePath)
		marshaleYaml, err := yaml.Marshal(namespace)

		if err != nil {
			log.Fatal(err)
		}

		writeErr := ioutil.WriteFile(yamlName, marshaleYaml, 0644)

		if writeErr != nil {
			log.Fatal(writeErr)
		}

		//save deployments
		SaveDeployments(Deployments.Items, namespacePath)

		//save statefulsets


		//eventually we will need to loop through the ss,ds,rs as well and save those
	}

}


func SaveDeployments(Deployments []v1.Deployment, path string) {
	//loop through each deployment and create a a deployment yaml
	for _, deploy := range Deployments{
		yamlName := fmt.Sprintf("%s/%s.deploy.yaml",  path, deploy.Name)

		//create kctl deployment struct
		deployment := KubeStruct{}


		deployment.ApiVersion = "apps/v1"
		deployment.Kind = "Deployment"

		deployment.Metadata = deploy.ObjectMeta
		deployment.Spec = deploy.Spec

		//save the deployment file
		marshalYaml, err:= yaml.Marshal(deployment)
		if err != nil {
			log.Panic(err)
		}
		writeErr := ioutil.WriteFile(yamlName, marshalYaml, 0644)
		if writeErr != nil {
			log.Fatal(writeErr)
		}
	}
}