package secretgen

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestGenerate(t *testing.T) {
	var expected []SecretManifest
	reader, err := os.Open("example/expected.yaml")
	defer reader.Close()
	d := yaml.NewDecoder(reader)
	for {
		var secret SecretManifest
		err = d.Decode(&secret)
		if err == io.EOF {
			break
		}
		check(err)
		expected = append(expected, secret)
	}

	var config Config
	configYaml, err := ioutil.ReadFile("example/generator.yaml")
	check(err)
	yaml.Unmarshal([]byte(configYaml), &config)
	check(err)

	got := Generate(config)
	assert.Equal(t, got, expected)
}
