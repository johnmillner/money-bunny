package yaml

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// getConfig deserialize yaml files into given object
func ParseYaml(path string, out interface{}) error {

	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read %s - please make sure it exsists", path)
	}

	err = yaml.Unmarshal(configFile, out)
	if err != nil {
		return fmt.Errorf("could not deserialize %s - please make sure to populate", path)
	}

	return nil
}
