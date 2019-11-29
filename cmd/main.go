package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/julienp/secretgen"
	"gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Expected exactly 2 arguments, got: %q", os.Args[1:])
	}
	var b []byte
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Could not read file %q, %v", os.Args[1], err)
	}
	var generatorConfig secretgen.Config
	err = yaml.Unmarshal(b, &generatorConfig)
	if err != nil {
		log.Fatalf("Could not unmarshal file %q, %v", os.Args[1], err)
	}

	s := secretgen.Generate(generatorConfig)

	for _, secret := range s {
		d, err := yaml.Marshal(secret)
		if err != nil {
			log.Fatalf("Could not marshal: %v", err)
		}
		fmt.Printf("---\n%s\n", string(d))
	}
}
