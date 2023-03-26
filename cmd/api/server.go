package api

import (
	"net/http"

	"github.com/OCD-Labs/KeyKeeper/internal/util"
)

type KeyKeeper struct{
	SwaggerSpec []byte
	Config util.Configs
}

func (app *KeyKeeper) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthcheck", app.ping)
	// Register the Swagger documentation handler
	mux.HandleFunc("/docs", app.serveDocs)
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(app.SwaggerSpec)
	})

	return mux
}
