package main

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

func saveFile(marshalYaml []byte, yamlName string) {
	writeErr := ioutil.WriteFile(yamlName, marshalYaml, 0644)
	if writeErr != nil {
		log.Error(writeErr)
	}
}

func deleteFile(path string) {
	err := os.Remove(path)
	if err != nil {
		log.Error(err)
	}
}

func CheckIfError(err error) {
	if err != nil {
		log.Error(err)
	}
}
