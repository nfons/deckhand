package main

import (
	"fmt"
	"github.com/ghodss/yaml" // this is better than regular yaml
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/robfig/cron.v2"
	"io/ioutil"
	"k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var clientset *kubernetes.Clientset

type DeckConfig struct {
	GitRepo        string `envconfig:"GIT_REPO" required:"true"`
	SyncInterval   string `default:"30s"`
	ClusterName    string `envconfig:"CLUSTER_NAME" default:"dev"`
	UseReplicaSets bool   `encconfig:"USE_REPLICA_SETS" default:"false"`
}

var deck_config DeckConfig
var createPath string

func main() {

	// set kubeconfig, probably will disable this later
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	err := envconfig.Process("deck", &deck_config)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Debugging for log purposes
	log.Println("Using Git Repo: " + deck_config.GitRepo)
	log.Println("Setting Cluster Name as: " + deck_config.ClusterName)
	log.Println("Setting Resync Interval at: " + deck_config.SyncInterval)

	createPath = filepath.Join("states", deck_config.ClusterName)

	// check if the cluster named folder exists
	if _, err := os.Stat(createPath); os.IsNotExist(err) {
		// need to create this path
		os.MkdirAll(createPath, 0777)
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset, to use for our api
	clientset, err = kubernetes.NewForConfig(config)

	// on init, get k8s state, this will get the latest
	log.Println("Syncing initial K8s State")
	GetKubernetesState(createPath)

	/*
		This code is removed because it doesn't play nice... I will need to work on this a bit more @ later time
		Will need to really look @ and see if its worth the effort to even do this
	*/

	c := cron.New()
	// Every SyncInterval, we will update to git master branch if there is an update needed
	c.AddFunc("@every "+deck_config.SyncInterval, sync)
	c.Start()
	WatchApis()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Syncs current k8s state with git state
func sync() {
	gitPullMaster()
	gitStatusCheck()
	gitPushMaster()

}

// initialize the remote git repo by pulling it. We will need to run this step first every time deck hand is init.
func gitPullMaster() {

}

// This function will check the status of the current git repo + against the remote
func gitStatusCheck() {

}

func gitPushMaster() {

}

// This function is the one that does the heavy lifting and actually gets the k8s state
// It is currently called on a Cron timer
func GetKubernetesState(createPath string) {
	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})

	if err != nil {
		log.Fatal(err)
	}
	// for each namespace, we will iterate and get deployments
	for _, val := range namespaces.Items {
		log.Printf("Getting State of Namespace: " + val.Name)
		Deployments, err := clientset.AppsV1().Deployments(val.Name).List(metav1.ListOptions{})
		// we should probably also get rs, ss, and ds as those would change as well.
		// we will get them, but leave them blank for now
		DaemonSets, _ := clientset.AppsV1().DaemonSets(val.Name).List(metav1.ListOptions{})
		StatefulSets, _ := clientset.AppsV1().StatefulSets(val.Name).List(metav1.ListOptions{})

		if err != nil {
			log.Fatal(err)
		}

		// create NS  dir if not exist
		namespacePath := filepath.Join(createPath, val.Name)

		// create the namespace path if it doesn't exist
		if _, exist := os.Stat(namespacePath); os.IsNotExist(exist) {
			os.MkdirAll(filepath.Join(namespacePath), 0777)
		}

		// save the namespace as well
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

		// save deployments
		SaveDeployments(Deployments.Items, namespacePath)

		// save statefulsets
		SaveStatefulset(StatefulSets.Items, namespacePath)

		SaveDaeomonSet(DaemonSets.Items, namespacePath)

		// Commenting out RS, since Deployments are just higher level RS
		if deck_config.UseReplicaSets == true {
			ReplicaSets, _ := clientset.AppsV1().ReplicaSets(val.Name).List(metav1.ListOptions{})
			SaveReplicaSets(ReplicaSets.Items, namespacePath)
		}

	}

}

/*
	Iterate through the deployments and save them to a file
*/
func SaveDeployments(Deployments []v1.Deployment, path string) {
	// loop through each deployment and create a a deployment yaml
	for _, deploy := range Deployments {
		SaveDeploy(deploy, path)
	}
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

/*
	Iterate through the deployments and save them to a file

*/
func SaveStatefulset(StatefulSets []v1.StatefulSet, path string) {
	// loop through each SS and create a a deployment yaml
	for _, deploy := range StatefulSets {
		SaveSS(deploy, path)
	}
}

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

func SaveDaeomonSet(daeomonset []v1.DaemonSet, path string) {
	// loop through each DS and create a a deployment yaml
	for _, deploy := range daeomonset {
		SaveDS(deploy, path)
	}
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

func SaveReplicaSets(rcs []v1.ReplicaSet, path string) {
	// loop through each RS and create a a deployment yaml
	for _, deploy := range rcs {
		SaveRS(deploy, path)
	}
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

func saveFile(marshalYaml []byte, yamlName string) {
	writeErr := ioutil.WriteFile(yamlName, marshalYaml, 0644)
	if writeErr != nil {
		log.Fatal(writeErr)
	}
}
