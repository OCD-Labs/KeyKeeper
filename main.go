package main

import (
	"log"
	"net/http"

	"github.com/OCD-Labs/KeyKeeper/cmd/api"
)

func main() {
	app, _  := api.NewkeyKeeper()

	log.Println("Starting server...")
	http.ListenAndServe(":8081", app.Routes())
}