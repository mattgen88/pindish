package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/AreaHQ/jsonhal"
	"github.com/mattgen88/pindish/models"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// BoardsResponse describes boards for user
type BoardsResponse struct {
	jsonhal.Hal
	Boards []models.PinterestBoard `json:"boards"`
}

// BoardsHandler Gets a list of user's boards
func (h *Handlers) BoardsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	u, err := getUserAccount(ctx, h.DB)

	if err != nil {
		log.WithFields(log.Fields{
			"msg": err,
		}).Warning("failed to get boards")

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		io.WriteString(w, "Failed to get user for boards")
		return
	}

	var boards []models.PinterestBoard
	if viper.GetBool("mock") {
		boards, err = getBoardsMock(*u.OAuth)
	} else {
		boards, err = getBoards(*u.OAuth)
	}

	j := &BoardsResponse{}
	j.Boards = boards
	j.SetLink("self", "/boards", "")

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

func getBoardsMock(o models.PinterestOAuthResponse) ([]models.PinterestBoard, error) {
	data, _ := ioutil.ReadFile("mocks/boards.json")
	var boards models.PinterestBoardResponse
	json.Unmarshal(data, &boards)
	return boards.Data, nil
}

func getBoards(o models.PinterestOAuthResponse) ([]models.PinterestBoard, error) {
	u := &url.URL{}
	u, _ = u.Parse("https://api.pinterest.com/v1/me/boards/")

	q := u.Query()
	q.Add("access_token", o.AccessToken)
	q.Add("fields", "id,link,url,color,board,metadata,counts,created_at,creator,image")

	u.RawQuery = q.Encode()

	log.WithField("url", u.String()).Info("fetching boards")

	response, err := netClient.Get(u.String())

	if err != nil {
		return nil, err
	}

	if response.StatusCode > 200 {
		return nil, fmt.Errorf("Bad status getting info on boards %d", response.StatusCode)
	}

	defer response.Body.Close()

	r := &models.PinterestBoardResponse{}
	json.NewDecoder(response.Body).Decode(r)

	return r.Data, nil

}
