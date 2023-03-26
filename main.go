package main

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/OCD-Labs/KeyKeeper/cmd/api"
)

//go:embed "docs/specs.yaml"
var embeddedSwaggerSpec []byte

func main() {
	app := api.KeyKeeper{
		SwaggerSpec: embeddedSwaggerSpec,
	}

	log.Println("Starting server...")
	http.ListenAndServe(":8081", app.Routes())
}
