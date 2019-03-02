package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mattgen88/pindish/models"
	"github.com/mattgen88/pindish/pinterest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type MyCustomClaims struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	FName string `json:"fname"`
	UName string `json:"uname"`
	Image string `json:"image"`
	jwt.StandardClaims
}

// CatchHandler handles the oauth catch
func (h *Handlers) CatchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	state := q.Get("state")
	code := q.Get("code")

	var err error
	var oauth *models.PinterestOAuthResponse

	// Get token from response
	if viper.GetBool("mock") {
		oauth, err = getAuthMock(state, code)
	} else {
		oauth, err = getAuth(state, code)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"state": state,
			"code":  code,
			"msg":   err,
		}).Warning("failed to get authorization token")

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		io.WriteString(w, "Failed to authorize with pinterest")
		return
	}

	var user *models.PinterestUser

	// Request info about user from pinterest
	if viper.GetBool("mock") {
		user, err = getUserMock(oauth)
	} else {
		user, err = pinterest.GetMe(oauth.AccessToken)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"oauth": oauth,
			"msg":   err,
		}).Warning("failed to get user data using oauth token")

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		io.WriteString(w, "Failed to authorize with pinterest")
		return
	}

	user.OAuth = oauth

	// Store user ID along with token in database
	err = createAccount(user, h.DB)
	if err != nil {
		log.WithFields(log.Fields{
			"user": user,
			"msg":  err,
		}).Warning("failed to create account for user")

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		io.WriteString(w, "Failed to create local account for pinterest user")
		return
	}

	mySigningKey := []byte(viper.GetString("signing_key"))

	// Create the Claims
	claims := MyCustomClaims{
		user.ID,
		user.OAuth.AccessToken,
		user.FirstName,
		user.UserName,
		user.Image["60x60"].URL,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    "pindish",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	if err != nil {
		log.WithFields(log.Fields{
			"user": user,
			"msg":  err,
		}).Warning("failed to issue jwt")

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		io.WriteString(w, "Failed to start session")
		return
	}

	// Set jwt
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   ss,
		Path:    "/",
		Expires: time.Now().Add(time.Hour * 24),
	})

	http.Redirect(w, r, fmt.Sprintf("%s", viper.GetString("frontend_uri")), http.StatusTemporaryRedirect)
	return
}

// getAuth will make a request to the pinterest API to get an oauth token for requests
func getAuth(state, code string) (*models.PinterestOAuthResponse, error) {

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
		return nil, fmt.Errorf("Bad status getting token for user %d", response.StatusCode)
	}

	defer response.Body.Close()

	oauth := &models.PinterestOAuthResponse{}
	json.NewDecoder(response.Body).Decode(oauth)

	return oauth, nil
}

func getAuthMock(state, code string) (*models.PinterestOAuthResponse, error) {
	data, _ := ioutil.ReadFile("mocks/oauth.json")
	var oauth models.PinterestOAuthResponse
	json.Unmarshal(data, &oauth)
	return &oauth, nil
}

func getUserMock(o *models.PinterestOAuthResponse) (*models.PinterestUser, error) {
	data, _ := ioutil.ReadFile("mocks/user.json")
	var user models.PinterestUserResponse
	json.Unmarshal(data, &user)
	return &user.Data, nil
}

func createAccount(u *models.PinterestUser, db *sql.DB) error {
	id, _ := strconv.Atoi(u.ID)

	rows, err := db.Query(`
		INSERT INTO users(
			id,
			first_name,
			username,
			url,
			image,
			token
		)
		VALUES(
			$1, $2, $3, $4, $5, $6
		) ON CONFLICT (id) DO NOTHING`,
		id,
		u.FirstName,
		u.UserName,
		u.URL,
		u.Image["60x60"].URL,
		u.OAuth.AccessToken,
	)
	if err != nil {
		log.WithField("id", id).WithField("msg", err).Warn("failed to insert user into database")
	}
	defer rows.Close()

	return err
}
