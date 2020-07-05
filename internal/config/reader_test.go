package config

import (
	"testing"
)

type Test struct {
	Test1 string
	Test2 string
}

func TestGetYamlConfig(t *testing.T) {
	test := Test{}
	err := GetYamlConfig("../../testdata/test.yaml", &test)

	if err != nil {
		t.Errorf("failed due to %s", err)
	}
	if "value1" != test.Test1 {
		t.Errorf("%s was not the expected key", test.Test1)
	}
	if "value2" != test.Test2 {
		t.Errorf("%s was not the expected secret", test.Test2)
	}
}

func TestGetYamlConfigNoFile(t *testing.T) {
	test := Test{}
	err := GetYamlConfig("../../testdata/does-not-exsist.yaml", &test)

	if err == nil || test != (Test{}) {
		t.Errorf("should not be able to load this file")
	}
}

func TestGetYamlConfigBadFile(t *testing.T) {
	test := Test{}
	err := GetYamlConfig("../../testdata/bad-file.yaml", &test)

	if err == nil || test != (Test{}) {
		t.Log(err)
		t.Log(test, Test{})
		t.Log(test != (Test{}))
		t.Errorf("should not be able to load this file")
	}
}
