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
func (h *Handlers) HomeHandler(w http.ResponseWriter, r *http.Request) {
	j := &HomeResponse{}
	j.SetLink("self", "/", "")
	j.SetLink("auth", "/auth", "Authorize")
	j.SetLink("catch", "/catch", "Catch endpoint from oauth process")
	j.SetLink("boards", "/boards", "List authorized user's boards")
	j.SetLink("search recipes", "/recipes", "Search for recipes by name/description")
	j.SetLink("recipes", "/recipes/board/{id}", "List recipes found on a board")
	j.SetLink("search ingredients", "/ingredients", "Search for recipes by ingredients")

	jsonResponse, err := json.Marshal(j)
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		io.WriteString(w, "Failed to build response")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "application/hal+json")
	w.Write(jsonResponse)
	return
}
