package main

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/OCD-Labs/KeyKeeper/cmd/api"
	"github.com/OCD-Labs/KeyKeeper/internal/util"
)

//go:embed "docs/specs.yaml"
var embeddedSwaggerSpec []byte

func main() {
	config, err := util.ParseConfigs("./")
	if err != nil {
		log.Fatalf("failed to parse configurations: %v", err)
	}

	app := api.KeyKeeper{
		SwaggerSpec: embeddedSwaggerSpec,
		Config: config,
	}

	log.Println("Starting server...")
	http.ListenAndServe(":8081", app.Routes())
}
