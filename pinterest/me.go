package pinterest

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mattgen88/pindish/models"
	log "github.com/sirupsen/logrus"
)

// GetMe will make a request to get information about the user who we just authorized
// This is necessary to create a database entry to associate data with the user
func GetMe(token string) (*models.PinterestUser, error) {
	u := &url.URL{}
	u, _ = u.Parse("https://api.pinterest.com/v1/me/")

	q := u.Query()
	q.Add("access_token", token)
	q.Add("fields", "first_name,id,last_name,url,image,username")

	u.RawQuery = q.Encode()

	log.WithField("url", u.String()).Info("fetching user")

	response, err := netClient.Get(u.String())

	if err != nil {
		return nil, err
	}

	if response.StatusCode > 200 {
		return nil, fmt.Errorf("Bad status getting info on user %d", response.StatusCode)
	}

	defer response.Body.Close()

	r := &models.PinterestUserResponse{}
	json.NewDecoder(response.Body).Decode(r)

	user := r.Data
	return &user, nil
}
