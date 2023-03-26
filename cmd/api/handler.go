package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/spec"
	"sigs.k8s.io/yaml"
)

func (app *KeyKeeper) ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}

func (app *KeyKeeper) serveDocs(w http.ResponseWriter, r *http.Request) {
	// Load the Swagger spec from the YAML file
	specJson, err := yaml.YAMLToJSON(app.SwaggerSpec)
	if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}

	var specDoc spec.Swagger
	if err := json.Unmarshal(specJson, &specDoc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
	}

	// Serve the Swagger UI
	uiHandler := middleware.SwaggerUI(middleware.SwaggerUIOpts{
			SpecURL:  "/swagger.json",
			BasePath: "/",
	}, nil)
	uiHandler.ServeHTTP(w, r)
}