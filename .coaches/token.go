//go:build ignore

// go run token.go -c coach-nigel.yaml | pbcopy
package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"os"

	"sigs.k8s.io/yaml"
)

type (
	team struct {
		Name string
		Link string
	}

	coach struct {
		Name  string
		Path  string
		Teams []team
	}
)

var yamlFile = flag.String("c", "", "yaml file to serialize")

func main() {
	flag.Parse()
	printToken(getCoach())
}

func getCoach() coach {
	f, err := os.ReadFile(*yamlFile)
	if err != nil {
		log.Fatal(err)
	}

	var c coach
	err = yaml.Unmarshal(f, &c)
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func printToken(c coach) {
	enc := json.NewEncoder(
		base64.NewEncoder(
			base64.StdEncoding,
			os.Stdout,
		),
	)

	err := enc.Encode(c)
	if err != nil {
		log.Fatal(err)
	}
}
