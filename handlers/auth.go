package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	uuid "github.com/nu7hatch/gouuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// AuthHandler redirects a user to the pinterest oauth
func (h *Handlers) AuthHandler(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewV4()
	if err != nil {
		log.Fatal("Could not generate uuid")
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		io.WriteString(w, "Could not generate state for oauth process")
		return
	}

	u := &url.URL{}
	u, _ = u.Parse("https://api.pinterest.com/oauth")

	q := u.Query()
	q.Add("response_type", "code")
	q.Add("client_id", viper.GetString("app_id"))
	q.Add("state", state.String())
	q.Add("scope", "read_public")
	q.Add("redirect_uri", fmt.Sprintf("https://%s/catch", r.Host))

	u.RawQuery = q.Encode()
	log.Info(q)

	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)

	return
}
