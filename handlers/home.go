package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/AreaHQ/jsonhal"
	log "github.com/sirupsen/logrus"
)

// HomeResponse is a response model for home
type HomeResponse struct {
	jsonhal.Hal
}

// HomeHandler handles the root api listing
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	j := &HomeResponse{}
	j.SetLink("self", "/", "")
	jsonResponse, err := json.Marshal(j)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(500)
		w.Header().Add("content-type", "text/plain")
		io.WriteString(w, "Failed to build response")
		return
	}

	w.WriteHeader(200)
	w.Header().Add("content-type", "application/hal+json")
	w.Write(jsonResponse)
	return
}
