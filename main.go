package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/warrensbox/terragrunt-versions-list/lib"
)

const (
	terraformAPIURL = "https://api.releases.hashicorp.com/v1/releases/terraform"
)

func main() {

	tfReleases, err := lib.GetTFReleases(terraformAPIURL) //get list of versions
	if err != nil {
		log.Fatalf("Encountered error while getting list of releases\nError: %v\n", err)
	}

	file, _ := json.MarshalIndent(tfReleases, "", " ")

	_ = ioutil.WriteFile("index.json", file, 0644)
}

type List struct {
	LastUpdated string
	Versions    []string `json:"Versions"`
}
