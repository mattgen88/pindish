package pinterest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mattgen88/pindish/models"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// GetAuth will make a request to the pinterest API to get an oauth token for requests
func GetAuth(state, code string) (*models.PinterestOAuthResponse, error) {

	u := &url.URL{}
	u, _ = u.Parse("https://api.pinterest.com/v1/oauth/token")

	q := u.Query()
	q.Add("grant_type", "authorization_code")
	q.Add("client_id", viper.GetString("app_id"))
	q.Add("client_secret", viper.GetString("app_secret"))
	q.Add("code", code)

	u.RawQuery = q.Encode()

	log.WithField("url", u.String()).Info("fetching token")

	response, err := netClient.Post(u.String(), "text/plain", bytes.NewReader([]byte{}))
	if err != nil {
		return nil, err
	}

	if response.StatusCode > 200 {
		if response.StatusCode == http.StatusTooManyRequests {
			return nil, ErrAPIOverLimit
		}
		return nil, fmt.Errorf("Bad status getting token for user %d", response.StatusCode)
	}

	defer response.Body.Close()

	oauth := &models.PinterestOAuthResponse{}
	json.NewDecoder(response.Body).Decode(oauth)

	return oauth, nil
}
