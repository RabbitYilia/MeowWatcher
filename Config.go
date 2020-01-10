package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

func ReadConfigFromFile(FileName string) map[string]interface{} {
	ConfigData, err := ioutil.ReadFile(FileName)
	if err != nil {
		log.Fatal(err)
	}
	var ConfigMap map[string]interface{}
	if err := json.Unmarshal(ConfigData, &ConfigMap); err != nil {
		log.Fatal(err)
	}
	return ConfigMap
}
