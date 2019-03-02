package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/AreaHQ/jsonhal"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// FavoritesResponse is our response to a favorites request
type FavoritesResponse struct {
	jsonhal.Hal
	Favorited bool `json:"favorited"`
}

type favoriteData struct {
	Favorited bool `json:"favorited"`
}

// FavoritesHandler Gets a list of user's boards
func (h *Handlers) FavoritesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var err error

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Warn(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		io.WriteString(w, "Failed to handle request	")
		return
	}

	var f favoriteData
	err = json.Unmarshal(body, &f)
	if err != nil {
		log.Warn(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		io.WriteString(w, "Failed to process body")
		return
	}

	j := &FavoritesResponse{
		Favorited: f.Favorited,
	}

	j.SetLink("self", "/favorite/board/"+id, "")

	jsonResponse, err := json.Marshal(j)
	if err != nil {
		log.Warn(err)
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

func setFavorite(boardID, userID string, value bool, db *sql.DB) error {
	rows, err := db.Query(`
		UPDATE owned_boards SET show=$1 WHERE user_id = $2 AND board_id = $3
	`, value, userID, boardID)
	if err != nil {
		log.WithField("msg", err).WithField("board", boardID).WithField("user", userID).WithField("value", value).Warn("Failed to set owned boards")
		return err
	}
	defer rows.Close()
	return nil
}
