package coinbase

import (
	"testing"
)

func TestGetCredentials(t *testing.T) {
	coinbase := getCredentials("../../testdata/coinbase-test.yaml")
	if "testKey" != coinbase.Auth.Key {
		t.Errorf("%s was not the expected key", coinbase.Auth.Key)
	}
	if "testSecret" != coinbase.Auth.Secret {
		t.Errorf("%s was not the expected secret", coinbase.Auth.Secret)
	}
}
