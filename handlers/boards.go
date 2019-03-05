package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/mattgen88/pindish/pinterest"

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
	Favorited   bool                   `json:"favorited"`
	LastUpdate  int                    `json:"last_update"`
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

	var boards []models.PinterestBoard
	getBoardsDb(getUserID(ctx), h.DB)
	if viper.GetBool("mock") {
		boards, err = getBoardsMock(token)
	} else {
		getBoardsDb(getUserID(ctx), h.DB)
		boards, err = pinterest.GetMyBoards(token)
	}

	j := &BoardsResponse{}

	var boardsResponse []Board

	for _, b := range boards {
		board := Board{
			Name:        b.Name,
			URL:         b.URL,
			Image:       b.Image,
			ID:          b.ID,
			Description: b.Description,
			Counts:      b.Counts,
		}
		putBoardDB(getUserID(ctx), b, h.DB)
		board.SetLink("recipes", fmt.Sprintf("/recipes/board/%s", board.ID), "recipes")
		boardsResponse = append(boardsResponse, board)
	}
	j.SetEmbedded("boards", jsonhal.Embedded(boardsResponse))
	j.SetEmbedded("me", jsonhal.Embedded(user))
	j.SetLink("self", "/boards", "")

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

func getBoardsMock(token string) ([]models.PinterestBoard, error) {
	data, _ := ioutil.ReadFile("mocks/boards.json")
	var boards models.PinterestBoardResponse
	json.Unmarshal(data, &boards)
	return boards.Data, nil
}

func putBoardDB(userID string, m models.PinterestBoard, db *sql.DB) error {
	uid, _ := strconv.Atoi(userID)
	boardsRows, err := db.Query(`
	INSERT INTO boards (
		id, name, url, description, image
	)
	VALUES (
		$1, $2, $3, $4, $5
	) ON CONFLICT(id) DO NOTHING`, m.ID, m.Name, m.URL, m.Description, m.Image["60x60"].URL)
	if err != nil {
		log.WithField("msg", err).WithField("userid", uid).WithField("board", m).Warn("Failed to insert into boards")
		return err
	}
	defer boardsRows.Close()

	ownedRows, err := db.Query(`
	INSERT INTO owned_boards (
		user_id, board_id, show, last_update
	)
	VALUES (
		$1, $2, $3, $4
	) ON CONFLICT DO NOTHING`, uid, m.ID, false, time.Now().Unix())
	if err != nil {
		log.WithField("msg", err).WithField("userid", uid).WithField("board", m).Warn("Failed to insert into owned_boards")
		return err
	}
	defer ownedRows.Close()
	return nil
}

func getBoardsDb(userID string, db *sql.DB) ([]Board, error) {
	var boards []Board
	rows, err := db.Query(`
		SELECT owned_boards.show, owned_boards.last_update, boards.id, boards.name, boards.url, boards.description, boards.image FROM owned_boards
		INNER JOIN boards on (owned_boards.board_id = boards.id)
		WHERE owned_boards.user_id = $1
	`, userID)
	if err != nil {
		log.WithField("uid", userID).Warn("Could not select boards")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var favorited bool
		var lastUpdate int
		var id int
		var name string
		var url string
		var description string
		var image string
		if err := rows.Scan(&favorited, &lastUpdate, &id, &name, &url, &description, &image); err != nil {
			log.WithField("userID", userID).WithField("msg", err).Warn("Error scanning")
			continue
		}
		spew.Dump(name)
		b := Board{
			Favorited:   favorited,
			LastUpdate:  lastUpdate,
			ID:          strconv.Itoa(id),
			Name:        name,
			URL:         url,
			Description: description,
			Image:       models.PinterestImages{"60x60": models.PinterestImage{URL: image, Height: 60, Width: 60}},
		}
		spew.Dump(b)
		boards = append(boards, b)
	}
	return boards, nil
}
