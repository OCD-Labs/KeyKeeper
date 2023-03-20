package api

import "net/http"

func (app *keyKeeper) ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world"))
}
