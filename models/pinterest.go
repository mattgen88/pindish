package models

// PinterestOAuthResponse describes a /v1/oauth/token response
type PinterestOAuthResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	Scope       []string `json:"scope"`
}

// PinterestUserResponse describes a /v1/me/ response
type PinterestUserResponse struct {
	Data PinterestUser `json:"data"`
}

// PinterestUser describes a pinterest user
type PinterestUser struct {
	ID        string          `json:"id"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	UserName  string          `json:"username"`
	URL       string          `json:"url"`
	Image     PinterestImages `json:"image"`
	OAuth     *PinterestOAuthResponse
}

// PinterestImages describes images
type PinterestImages map[string]PinterestImage

// PinterestImage describes a single image
type PinterestImage struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

// PinterestBoardResponse describes a /v1/me/boards response
type PinterestBoardResponse struct {
	Data []PinterestBoard `json:"data"`
}

// PinterestBoard describes a pinterest board
type PinterestBoard struct {
	Name        string          `json:"name"`
	URL         string          `json:"url"`
	Image       PinterestImages `json:"image"`
	Counts      PinterestCount  `json:"counts"`
	ID          string          `json:"id"`
	Description string          `json:"description"`
}

// PinterestCount describes a pinterest board count
type PinterestCount struct {
	Pins          int `json:"pins"`
	Collaborators int `json:"collaborators"`
	Followers     int `json:"followers"`
	Saves         int `json:"saves"`
	Comments      int `json:"comments"`
}

//PinterestPinsResponse describes /v1/me/boards/{id}/pins/ response
type PinterestPinsResponse struct {
	Data []PinterestPins `json:"data"`
}

// PinterestPins describes the pins on a board
type PinterestPins struct {
	Attribution  string               `json:"attribution"`
	Creator      PinterestCreator     `json:"creator"`
	URL          string               `json:"url"`
	OriginalLink string               `json:"original_link"`
	Note         string               `json:"note"`
	Color        string               `json:"color"`
	Link         string               `json:"link"`
	Board        PinterestBoard       `json:"board"`
	Image        PinterestImages      `json:"image"`
	Counts       PinterestCount       `json:"count"`
	ID           string               `json:"id"`
	Metadata     PinterestPinMetadata `json:"metadata"`
}

// PinterestPinMetadata describes a pin's metadata
type PinterestPinMetadata struct {
	Article *MetadataArticle `json:"article"`
	Link    *MetadataLink    `json:"link"`
	Recipe  *MetadataRecipe  `json:"recipe"`
}

// MetadataArticle describes the article
type MetadataArticle struct {
	PublishedAt string `json:"published_at"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Authors     []struct {
		Name string `json:"name"`
	} `json:"authors"`
}

// MetadataLink describes the link
type MetadataLink struct {
	Locale      string `json:"locale"`
	Title       string `json:"title"`
	SiteName    string `json:"site_name"`
	Description string `json:"description"`
	Favicon     string `json:"favicon"`
}

// MetadataRecipe describes a recipe
type MetadataRecipe struct {
	Servings struct {
		Serves  string `json:"serves"`
		Summary string `json:"summary"`
	} `json:"servings"`
	Name                 string               `json:"name"`
	IngredientCategories []IngredientCategory `json:"ingredients"`
}

// IngredientCategory describes an ingredient category
type IngredientCategory struct {
	Category    string       `json:"category"`
	Ingredients []Ingredient `json:"ingredients"`
}

// Ingredient describes an ingredient
type Ingredient struct {
	Amount string `json:"amount"`
	Name   string `json:"name"`
}

// PinterestCreator describes a creator
type PinterestCreator struct {
	URL       string `json:"url"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	ID        string `json:"id"`
}
