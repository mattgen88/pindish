package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/mattgen88/pindish/pinterest"

	"github.com/AreaHQ/jsonhal"
	"github.com/gorilla/mux"
	"github.com/mattgen88/pindish/models"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// RecipesResponse describes boards for user
type RecipesResponse struct {
	jsonhal.Hal
}

// Recipe describes a recipe
type Recipe struct {
	jsonhal.Hal
	Image    models.PinterestImages `json:"image"`
	Link     string                 `json:"link"`
	Servings struct {
		Serves  string `json:"serves"`
		Summary string `json:"summary"`
	} `json:"servings"`
	Name        string              `json:"name"`
	Ingredients []models.Ingredient `json:"ingredients"`
}

// RecipesHandler Gets a list of user's boards
func (h *Handlers) RecipesHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	ctx := r.Context()
	token := getToken(ctx)

	vars := mux.Vars(r)
	id := vars["id"]

	if err != nil {
		log.WithFields(log.Fields{
			"msg": err,
		}).Warning("failed to get boards")

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Add("content-type", "text/plain")
		io.WriteString(w, "Failed to get user for boards")
		return
	}

	var pins []models.PinterestPins
	if viper.GetBool("mock") {
		pins, err = getPinsMock(token, id)
	} else {
		pins, err = pinterest.GetBoardPins(token, id)
	}

	j := &RecipesResponse{}
	j.SetLink("self", "/recipes/board/"+id, "")

	var recipes []Recipe

	for _, pin := range pins {
		if pin.Metadata.Recipe != nil {
			recipe := Recipe{
				Link:        pin.OriginalLink,
				Image:       pin.Image,
				Servings:    pin.Metadata.Recipe.Servings,
				Name:        pin.Metadata.Recipe.Name,
				Ingredients: pin.Metadata.Recipe.Ingredients,
			}
			recipes = append(recipes, recipe)
		}
	}

	j.SetEmbedded("recipes", jsonhal.Embedded(recipes))

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

func getPinsMock(token, id string) ([]models.PinterestPins, error) {
	data, _ := ioutil.ReadFile("mocks/pins.json")
	var pins models.PinterestPinsResponse
	json.Unmarshal(data, &pins)
	return pins.Data, nil
}

func getRecipes(id string, db *sql.DB) ([]models.PinterestPins, error) {
	rows, err := db.Query(`SELECT * FROM recipes WHERE board_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return nil, nil
}

func putRecipes(userid, boardid string, r RecipesResponse) error {
	// foreach recipes in the response
	// insert the recipe into the database
	// insert the relation table
	return nil
}
