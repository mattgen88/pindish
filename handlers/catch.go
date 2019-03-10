package handlers

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mattgen88/pindish/models"
	"github.com/mattgen88/pindish/pinterest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// PindishClaims are the custom claims for pindish
type PindishClaims struct {
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
	oauth, err = pinterest.GetAuth(state, code)

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
	user, err = pinterest.GetMe(oauth.AccessToken)

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

	mySigningKey := []byte(viper.GetString("signing_key"))

	// Create the Claims
	claims := PindishClaims{
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
