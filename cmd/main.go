package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/julienp/secretgen"
	"gopkg.in/yaml.v2"
)

func lookupEnvOrBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		if val == "" {
			return defaultVal
		}
		boolVal, err := strconv.ParseBool(val)
		if err != nil {
			return defaultVal
		}
		return boolVal
	}
	return defaultVal
}

func main() {
	fake := flag.Bool("fake", lookupEnvOrBool("SECRETGEN_FAKE", false), "fake sops by not invoking")
	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatalf("Expected exactly 1 argument, got: %q", flag.Args())
	}

	manifest := flag.Arg(0)

	var b []byte
	b, err := ioutil.ReadFile(manifest)
	if err != nil {
		log.Fatalf("Could not read file %q, %v", manifest, err)
	}
	var generatorConfig secretgen.Config
	err = yaml.Unmarshal(b, &generatorConfig)
	if err != nil {
		log.Fatalf("Could not unmarshal file %q, %v", manifest, err)
	}

	s := secretgen.Generate(generatorConfig, *fake)

	for _, secret := range s {
		d, err := yaml.Marshal(secret)
		if err != nil {
			log.Fatalf("Could not marshal: %v", err)
		}
		fmt.Printf("---\n%s\n", string(d))
	}
}
