package main

import (
	"crypto/tls"
	"fmt"
	"github.com/ghodss/yaml" // this is better than regular yaml
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"gopkg.in/robfig/cron.v2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/client"
	http2 "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	ssh2 "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var clientset *kubernetes.Clientset

const directory = "repo"

var deck_config DeckConfig
var createPath string
var auth transport.AuthMethod

type DeckConfig struct {
	GitRepo        string `envconfig:"GIT_REPO" required:"true"`
	SyncInterval   string `default:"30s"`
	ClusterName    string `envconfig:"CLUSTER_NAME" default:"dev"`
	UseReplicaSets bool   `encconfig:"USE_REPLICA_SETS" default:"false"`
	SSH_KEY        string `envconfig:"SSH_KEY"`
	KUBE_CONF      string `envconfig:"KUBE_CONF"`
	STORE_ALL      bool   `envconfig:"STORE_ALL" default:"false"`
	GitUser        string `split_words:"true"`
	GitPassword    string `split_words:"true"`
}

func main() {

	// Set Up
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	enverr := envconfig.Process("deck", &deck_config)
	if enverr != nil {
		log.Fatal(enverr.Error())
	}

	// Allow the passing in of kubeconf as a env var..really only useful for running this via docker outside cluster
	// We probably can get rid of this
	if deck_config.KUBE_CONF != "" {
		// create k8s conf path if it doesn't exist
		if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".kube")); os.IsNotExist(err) {
			// need to create this path
			os.MkdirAll(filepath.Join(os.Getenv("HOME"), ".kube"), 0777)
			file, openErr := os.Create(kubeconfig)
			if openErr != nil {
				log.Fatal("Cannot Open File " + openErr.Error())
			}
			defer file.Close()
			fmt.Fprintf(file, deck_config.KUBE_CONF)
		}
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Println(err.Error())
		// Try to use local config?
		// set kubeconfig, probably will disable this later
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}

	}
	// create the clientset, to use for our api
	clientset, err = kubernetes.NewForConfig(config)

	// Debugging for log purposes
	log.Info("Using Git Repo: " + deck_config.GitRepo)
	log.Info("Setting Cluster Name as: " + deck_config.ClusterName)
	log.Info("Setting Resync Interval at: " + deck_config.SyncInterval)

	createPath = filepath.Join(directory, "state", deck_config.ClusterName)

	// Clone the Git Repo

	if deck_config.GitPassword == "" {
		sshkey := deck_config.SSH_KEY
		signer, _ := ssh.ParsePrivateKey([]byte(sshkey))
		sshAuth := &ssh2.PublicKeys{User: "git", Signer: signer}

		// needed for known_host error during docker runs
		sshAuth.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		auth = sshAuth
	} else {

		// Get a custom client
		customClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 15 * time.Second, // 15 second timeout
			CheckRedirect: func(req *http.Request, via []*http.Request) error { // don't follow redirect
				return http.ErrUseLastResponse
			},
		}
		client.InstallProtocol("https", http2.NewClient(customClient))

		auth = &http2.BasicAuth{Username: deck_config.GitUser, Password: deck_config.GitPassword}
	}

	_, giterr := git.PlainClone(directory, false, &git.CloneOptions{
		URL:      deck_config.GitRepo,
		Progress: os.Stdout,
		Auth:     auth,
	})

	if giterr != nil {
		log.Fatal(giterr)
	}

	// pull master real quick to make sure we are tip top updated
	gitPullMaster()

	// check if the cluster named folder exists
	if _, err := os.Stat(createPath); os.IsNotExist(err) {
		// need to create this path
		os.MkdirAll(createPath, 0777)
	} else {
		// this path exists...But it could be old, so we need to delete that whole dir
		log.Info("Flushing old repo files")
		if os.RemoveAll(createPath+"/") != nil {
			log.Fatal("Could not clear old dir struct")
		}
		if os.MkdirAll(createPath, 0777) != nil {
			log.Fatal("Could not recreate dir struct after flusing")
		}
	}

	// on init, get k8s state, this will get the latest
	log.Info("Syncing initial K8s State")
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
	if gitStatusCheck() == false {
		log.Info("Status is dirty, Pushing new state to repo")
		gitPushMaster()
	} else {
		log.Println(" Status is clean, skipping push")
	}

}

// initialize the remote git repo by pulling it. We will need to run this step first every time deck hand is init.
func gitPullMaster() {
	log.Debug("Pulling From master repo")
	r, err := git.PlainOpen(directory)
	CheckIfError(err)
	worktree, err := r.Worktree()

	CheckIfError(err)

	pullerr := worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth:       auth,
		Progress:   os.Stdout,
	})

	if pullerr != nil {
		errstr := pullerr.Error()
		if errstr != "already up-to-date" {
			log.Info(errstr)
		} else {
			log.Info("Pull Already up-to-date")
		}
	}
}

// This function will check the status of the current git repo + against the remote
func gitStatusCheck() bool {
	r, err := git.PlainOpen(directory)
	CheckIfError(err)

	workTree, err := r.Worktree()

	CheckIfError(err)
	status, err := workTree.Status()
	CheckIfError(err)

	return status.IsClean()
}

func gitPushMaster() {
	r, err := git.PlainOpen(directory)
	CheckIfError(err)

	// add the outstanding files
	worktree, err := r.Worktree()

	CheckIfError(err)

	// Get the outstanding items
	status, _ := worktree.Status()

	for key := range status {
		worktree.Add(key)
	}
	if len(status) == 0 {
		log.Error("Status len of 0 was returned as clean: ", status.IsClean())
		return
	} else {
		log.Info("Changes detected in files:")
		for key := range status {
			log.Info(key)
		}
	}

	// Commit the current state to git
	commitMsg := fmt.Sprintf("K8s State @ %s ", time.Now().Format(time.RFC1123))
	commit, err := worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name: "Deck Hand",
			When: time.Now(),
		},
	})
	CheckIfError(err)

	_, err = r.CommitObject(commit)

	CheckIfError(err)

	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
	})

	CheckIfError(err)
	log.Info("Latest changes pushed to repo")
}

// This function is the one that does the heavy lifting and actually gets the k8s state
func GetKubernetesState(createPath string) {
	namespaces, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})

	if err != nil {
		log.Fatal(err)
	}
	// for each namespace, we will iterate and get deployments
	for _, val := range namespaces.Items {
		log.Printf("Getting State of Namespace: " + val.Name)
		// Deployments, err := clientset.AppsV1().Deployments(val.Name).List(metav1.ListOptions{})
		// // // we should probably also get rs, ss, and ds as those would change as well.
		// // // we will get them, but leave them blank for now
		// DaemonSets, _ := clientset.AppsV1().DaemonSets(val.Name).List(metav1.ListOptions{})
		// StatefulSets, _ := clientset.AppsV1().StatefulSets(val.Name).List(metav1.ListOptions{})

		if err != nil {
			log.Fatal(err)
		}

		// create NS  dir if not exist
		namespacePath := filepath.Join(createPath, val.Name)

		// create the namespace path if it doesn't exist
		if _, exist := os.Stat(namespacePath); os.IsNotExist(exist) {
			if os.MkdirAll(filepath.Join(namespacePath), 0777) != nil {
				log.Fatal("Can't recreate namespace path")
			}
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
			log.Println(err)
		}

		writeErr := ioutil.WriteFile(yamlName, marshaleYaml, 0644)

		if writeErr != nil {
			log.Println(writeErr)
		}

	}

}
