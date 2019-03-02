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
	Name        string              `json:"name"`
	URL         string              `json:"url"`
	Image       PinterestImages     `json:"image"`
	Counts      PinterestBoardCount `json:"counts"`
	ID          int                 `json:"id"`
	Description string              `json:"description"`
}

// PinterestBoardCount describes a pinterest board count
type PinterestBoardCount struct {
	Pins          int `json:"pins"`
	Collaborators int `json:"collaborators"`
	Followers     int `json:"followers"`
}
