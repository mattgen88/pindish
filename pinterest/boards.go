package pinterest

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mattgen88/pindish/models"
	log "github.com/sirupsen/logrus"
)

// GetMyBoards returns a list of your boards
func GetMyBoards(token string) ([]models.PinterestBoard, error) {
	u := &url.URL{}
	u, _ = u.Parse("https://api.pinterest.com/v1/me/boards/")

	q := u.Query()
	q.Add("access_token", token)
	q.Add("fields", "id,name,url,counts,description,image")

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
