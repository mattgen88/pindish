package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dgrijalva/jwt-go"
	"github.com/mattgen88/pindish/models"
	uuid "github.com/nu7hatch/gouuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// AuthHandler redirects a user to the pinterest oauth
func (h *Handlers) AuthHandler(w http.ResponseWriter, r *http.Request) {
	state, err := uuid.NewV4()
	if err != nil {
		log.Fatal("Could not generate uuid")
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		io.WriteString(w, "Could not generate state for oauth process")
		return
	}

	u := &url.URL{}
	u, _ = u.Parse("https://api.pinterest.com/oauth")

	q := u.Query()
	q.Add("response_type", "code")
	q.Add("client_id", viper.GetString("app_id"))
	q.Add("state", state.String())
	q.Add("scope", "read_public")
	q.Add("redirect_uri", fmt.Sprintf("https://%s/catch", r.Host))

	u.RawQuery = q.Encode()
	log.Info(q)

	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)

	return
}

type key int

// CtxUserID is for looking up the user id from context
const CtxUserID key = iota

func getUserAccount(ctx context.Context, db *sql.DB) (*models.PinterestUser, error) {
	id := ctx.Value(CtxUserID).(string)
	var fname, uname, image, url, token string
	err := db.QueryRow(`SELECT first_name, username, image, url, token FROM users WHERE id = $1`, id).Scan(&fname, &uname, &image, &url, &token)
	if err != nil {
		log.WithField("id", id).Warn("Failed to query account")
		return nil, err
	}
	return &models.PinterestUser{
		ID:        id,
		FirstName: fname,
		UserName:  uname,
		Image: models.PinterestImages{
			"60x60": models.PinterestImage{
				URL:    image,
				Height: 60,
				Width:  60,
			},
		},
		OAuth: &models.PinterestOAuthResponse{
			AccessToken: token,
		},
	}, nil
}

// AuthRequired is middleware that handles auth checking
func AuthRequired(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		if err != nil {
			log.WithField("cookies", c).Warn("No token cookie")
			// reject request
			w.Header().Add("content-type", "text/plain")
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, "Authorization required")
			return
		}
		tokenString := c.Value

		// Parse takes the token string and a function for looking up the key. The latter is especially
		// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
		// head of the token to identify which key to use, but the parsed token (head and claims) is provided
		// to the callback, providing flexibility.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Warn("token cookie not signed correctly")
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(viper.GetString("signing_key")), nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			ctx := context.Background()
			ctx = context.WithValue(ctx, CtxUserID, claims["id"])
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// reject request
		log.Warn("No valid token cookie")

		w.Header().Add("content-type", "text/plain")
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Authorization required")
	})
}
