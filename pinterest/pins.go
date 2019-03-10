package pinterest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mattgen88/pindish/models"
	log "github.com/sirupsen/logrus"
)

// GetBoardPins gets the pins for a board
func GetBoardPins(token, id string) ([]models.PinterestPins, error) {
	u := &url.URL{}
	u, _ = u.Parse(fmt.Sprintf("https://api.pinterest.com/v1/boards/%s/pins/", id))

	q := u.Query()
	q.Add("access_token", token)
	q.Add("fields", "id,link,note,url,metadata,creator,original_link,image")

	u.RawQuery = q.Encode()

	log.WithField("url", u.String()).WithField("pin", id).Info("fetching pins")

	response, err := netClient.Get(u.String())

	if err != nil {
		return nil, err
	}

	if response.StatusCode > 200 {
		if response.StatusCode == http.StatusTooManyRequests {
			return nil, ErrAPIOverLimit
		}
		return nil, fmt.Errorf("Bad status getting info on pins %d", response.StatusCode)
	}

	defer response.Body.Close()

	r := &models.PinterestPinsResponse{}
	json.NewDecoder(response.Body).Decode(r)

	return r.Data, nil
}
