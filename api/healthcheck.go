package api

import (
	"encoding/json"
	"net/http"
)

func (app *KeyKeeper) ping(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"database_status": "healthy",
		"service_status":  "healthy",
	}

	buf, err := json.Marshal(health)
	if err != nil {
		http.Error(w, "failed to serialize health data", http.StatusInternalServerError)
	}

	w.Write([]byte(buf))
}
