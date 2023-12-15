package main

import (
	"log"
	"os"

	"github.com/cnrancher/autok3s/pkg/client/generator"
	"github.com/cnrancher/autok3s/pkg/server/schema"
	"github.com/rancher/apiserver/pkg/builtin"
	"github.com/rancher/apiserver/pkg/types"
)

const (
	outputDir = "./pkg/client/generated"
)

func main() {
	os.Unsetenv("GOPATH")
	version := "v1"
	s := types.EmptyAPISchemas().MustAddSchemas(builtin.Schemas)
	if err := generator.GenerateClient(schema.InitSchema(s), nil, outputDir, version); err != nil {
		log.Fatal(err)
	}
}
