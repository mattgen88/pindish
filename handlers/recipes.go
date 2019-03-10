package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/mattgen88/pindish/pinterest"

	"github.com/AreaHQ/jsonhal"
	"github.com/gorilla/mux"
	"github.com/mattgen88/pindish/models"
	log "github.com/sirupsen/logrus"
)

// RecipesResponse describes boards for user
type RecipesResponse struct {
	jsonhal.Hal
	Error error `json:"error,omitempty"`
}

// Recipe describes a recipe
type Recipe struct {
	jsonhal.Hal
	ID       int                    `json:"id"`
	Image    models.PinterestImages `json:"image"`
	Link     string                 `json:"link"`
	Servings struct {
		Serves  string `json:"serves"`
		Summary string `json:"summary"`
	} `json:"servings"`
	Name        string                      `json:"name"`
	Ingredients []models.IngredientCategory `json:"ingredients"`
}

// RecipesHandler Gets a list of user's boards
func (h *Handlers) RecipesHandler(w http.ResponseWriter, r *http.Request) {

	var err error
	ctx := r.Context()
	token := getToken(ctx)

	vars := mux.Vars(r)
	id := vars["id"]

	j := &RecipesResponse{}
	j.SetLink("self", "/recipes/board/"+id, "")

	var recipes []Recipe
	recipes, err = getRecipes(id, h.DB)
	if err != nil || len(recipes) == 0 {

		var pins []models.PinterestPins
		pins, err = pinterest.GetBoardPins(token, id)
		if err != nil {
			j.Error = err
		}

		for _, pin := range pins {
			if pin.Metadata.Recipe != nil {
				recipeID, _ := strconv.Atoi(pin.ID)
				recipe := Recipe{
					ID:          recipeID,
					Link:        pin.OriginalLink,
					Image:       pin.Image,
					Servings:    pin.Metadata.Recipe.Servings,
					Name:        pin.Metadata.Recipe.Name,
					Ingredients: pin.Metadata.Recipe.IngredientCategories,
				}
				recipes = append(recipes, recipe)
			}
		}
		putRecipes(getUserID(ctx), id, pins, h.DB)
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

func getRecipes(id string, db *sql.DB) ([]Recipe, error) {
	var recipes []Recipe
	recipeRows, err := db.Query(`
		SELECT recipes.id, recipes.name, recipes.url, recipes.serves, recipes.serving_summary, recipes.image FROM recipes
		INNER JOIN board_recipes on (board_recipes.recipe_id = recipes.id)
		WHERE board_recipes.board_id = $1
	`, id)
	if err != nil {
		log.Warn(err)
		return nil, err
	}
	defer recipeRows.Close()
	for recipeRows.Next() {
		r := Recipe{}
		var image string
		err = recipeRows.Scan(&r.ID, &r.Name, &r.Link, &r.Servings.Serves, &r.Servings.Summary, &image)
		if err != nil {
			log.Warn(err)
			return nil, err
		}
		r.Image = models.PinterestImages{"original": models.PinterestImage{URL: image}}
		ingredientRows, err := db.Query(`
			SELECT ingredients.category, ingredients.name, recipe_ingredients.quantity FROM ingredients
			INNER JOIN recipe_ingredients on (ingredients.id = recipe_ingredients.ingredient_id)
			WHERE recipe_ingredients.recipe_id = $1
		`, r.ID)
		if err != nil {
			log.Warn(err)
			return nil, err
		}
		defer ingredientRows.Close()
		var categoryIngredientMap map[string][]models.Ingredient
		categoryIngredientMap = make(map[string][]models.Ingredient)
		for ingredientRows.Next() {
			var category string
			i := models.Ingredient{}
			ingredientRows.Scan(&category, &i.Name, &i.Amount)
			if _, ok := categoryIngredientMap[category]; !ok {
				categoryIngredientMap[category] = []models.Ingredient{i}
			} else {
				categoryIngredientMap[category] = append(categoryIngredientMap[category], i)
			}
		}
		for c, i := range categoryIngredientMap {
			r.Ingredients = append(r.Ingredients, models.IngredientCategory{Category: c, Ingredients: i})
		}
		recipes = append(recipes, r)
	}
	return recipes, nil
}

func putRecipes(userid, boardid string, pins []models.PinterestPins, db *sql.DB) error {
	for _, p := range pins {
		if p.Metadata.Recipe == nil || p.Metadata.Article == nil {
			continue
		}
		db.Exec(`
			INSERT INTO recipes
			(id, name, url, serves, serving_summary, image)
			VALUES
			($1, $2, $3, $4, $5)
		`,
			p.ID,
			p.Metadata.Article.Name,
			p.URL,
			p.Metadata.Recipe.Servings.Serves,
			p.Metadata.Recipe.Servings.Summary,
			p.Image,
		)

		db.Exec(`
			INSERT INTO board_recipes
			(board_id, recipe_id)
			VALUES
			($1, $2)
			ON CONFLICT (board_id, recipe_id) DO NOTHING`, boardid, p.ID)

		for _, ingCategory := range p.Metadata.Recipe.IngredientCategories {
			for _, i := range ingCategory.Ingredients {
				putRecipe(p.ID, i, ingCategory.Category, db)
			}
		}
	}
	return nil
}

func putRecipe(boardID string, i models.Ingredient, category string, db *sql.DB) {
	var ingredientID int
	ingredientRow := db.QueryRow(`
		SELECT ingredients.id from ingredients WHERE name=$1
	`, i.Name)
	err := ingredientRow.Scan(&ingredientID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Warn(err)
		} else {
			// insert
			err := db.QueryRow(`
				INSERT INTO ingredients (category, name)
				VALUES ($1, $2)
				ON CONFLICT (category, name) DO UPDATE SET category=$1, name=$2
				RETURNING ingredients.id
			`, category, i.Name).Scan(&ingredientID)
			if err != nil {
				log.Warn(err)
				return
			}
		}
	}
	recipeIngredientRows, err := db.Query(`
		INSERT INTO recipe_ingredients
		(recipe_id, ingredient_id, quantity)
		VALUES
		($1, $2, $3)
	`, boardID, ingredientID, i.Amount)

	defer recipeIngredientRows.Close()
}
