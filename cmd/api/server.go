package api

import "net/http"

type KeyKeeper struct{
	SwaggerSpec []byte
}

func (app *KeyKeeper) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.ping)
	// Register the Swagger documentation handler
	mux.HandleFunc("/docs", app.serveDocs)
	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write(app.SwaggerSpec)
	})

	return mux
}
