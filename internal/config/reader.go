package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// GetYamlConfig deserialize yaml files into given object
func GetYamlConfig(path string, out interface{}) error {

	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read %s - please make sure it exsists", path)
	}

	err = yaml.UnmarshalStrict(configFile, out)
	if err != nil {
		return fmt.Errorf("could not deserialize %s - please make sure to populate", path)
	}

	return nil
}
