package handlers

import (
	"net/http"

	"github.com/AreaHQ/jsonhal"
)

// HomeResponse is a response model for home
type HomeResponse struct {
	jsonhal.Hal
}

// HomeHandler handles the root api listing
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	j := &HomeResponse{}
	j.SetLink("self", "/", "")
	return
}
