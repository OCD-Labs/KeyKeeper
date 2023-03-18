package api

import "net/http"

type keyKeeper struct {}

func NewkeyKeeper() (*keyKeeper, error) {
	return &keyKeeper{}, nil
}

func (app *keyKeeper) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.ping)

	return mux
}
