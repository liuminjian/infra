package base

import (
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func GetConfig(configFile string, value interface{}) error {
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Errorf("read yaml file fail:%s", err.Error())
		return err
	}
	err = yaml.Unmarshal(yamlFile, value)
	if err != nil {
		log.Errorf("Unmarshal yaml file fail:%s", err.Error())
		return err
	}

	if err := ValidateStruct(value); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
