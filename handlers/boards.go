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

// User describes a user
type User struct {
	Name     string `json:"name"`
	Image    string `json:"image"`
	Username string `json:"username"`
}

// BoardsResponse describes boards for user
type BoardsResponse struct {
	jsonhal.Hal
}

// Board describes a board resource
type Board struct {
	jsonhal.Hal
	Name        string                 `json:"name"`
	URL         string                 `json:"url"`
	Image       models.PinterestImages `json:"image"`
	Counts      models.PinterestCount  `json:"counts"`
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
}

// BoardsHandler Gets a list of user's boards
func (h *Handlers) BoardsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	token := getToken(ctx)

	user := User{
		Name:     getUserFName(ctx),
		Image:    getUserImage(ctx),
		Username: getUserUName(ctx),
	}

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
		boards, err = getBoardsMock(token)
	} else {
		boards, err = getBoards(token)
	}

	j := &BoardsResponse{}

	var boardsResponse []Board
	// @TODO: Fix this to add an embed one at a time instead of as an array, giving an id as the key
	for _, b := range boards {
		board := Board{
			Name:        b.Name,
			URL:         b.URL,
			Image:       b.Image,
			ID:          b.ID,
			Description: b.Description,
			Counts:      b.Counts,
		}
		board.SetLink("recipes", fmt.Sprintf("/recipes/board/%s", board.ID), "recipes")
		boardsResponse = append(boardsResponse, board)
	}
	j.SetEmbedded("boards", jsonhal.Embedded(boardsResponse))
	j.SetEmbedded("me", jsonhal.Embedded(user))
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

func getBoardsMock(token string) ([]models.PinterestBoard, error) {
	data, _ := ioutil.ReadFile("mocks/boards.json")
	var boards models.PinterestBoardResponse
	json.Unmarshal(data, &boards)
	return boards.Data, nil
}

func getBoards(token string) ([]models.PinterestBoard, error) {
	u := &url.URL{}
	u, _ = u.Parse("https://api.pinterest.com/v1/me/boards/")

	q := u.Query()
	q.Add("access_token", token)
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
