package base

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

func GetConfig(configFile string, value interface{}) {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("read yaml file fail:%s", err.Error())
	}
	err = yaml.Unmarshal(yamlFile, value)
	if err != nil {
		log.Fatalf("Unmarshal yaml file fail:%s", err.Error())
	}

	if err := ValidateStruct(value); err != nil {
		log.Fatal(err)
	}
}
