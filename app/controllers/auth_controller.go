package controllers

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/bentranter/urlscraper/app/models"

	"github.com/bentranter/go-seatbelt"
	"github.com/digitalocean/godo"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

const doUserEndpoint = "https://api.digitalocean.com/v2/account"

var oauth2Config *oauth2.Config

func init() {
	clientID := "240c4117956682bd397101dc26ba01dedc5aabcd42bbf7948f81894180d5a6da"
	clientSecret := "e499b6217be1074868eba76b0de3038c20726149f9fee7bd69933c7694b596c4"
	redirectURL := "http://localhost:3000/auth/digitalocean/callback"

	if cid := os.Getenv("CLIENT_ID"); cid != "" {
		clientID = cid
	}
	if cs := os.Getenv("CLIENT_SECRET"); cs != "" {
		clientSecret = cs
	}
	if ru := os.Getenv("REDIRECT_URL"); ru != "" {
		redirectURL = ru
	}

	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://cloud.digitalocean.com/v1/oauth/authorize",
			TokenURL: "https://cloud.digitalocean.com/v1/oauth/token",
		},
	}
}

// AuthIndex renders the login page.
func AuthIndex(c seatbelt.Context) error {
	return c.NoContent()
}

// AuthRedirect redirects the user to Google's OAuth consent page.
func AuthRedirect(c seatbelt.Context) error {
	return c.Redirect(oauth2Config.AuthCodeURL("state"))
}

// AuthCallback redirects the user to Google's OAuth consent page.
func AuthCallback(c seatbelt.Context) error {
	token, err := oauth2Config.Exchange(context.Background(), c.FormValue("code"))
	if err != nil {
		return err
	}

	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
	resp, err := client.Get(doUserEndpoint)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Debug().Str("google_user_data", string(data)).Msg("google user data")

	account := &struct {
		Account *godo.Account `json:"account"`
	}{}
	if err := json.Unmarshal(data, account); err != nil {
		return err
	}

	user := &models.User{
		Email: account.Account.Email,
	}
	if _, err := models.FindOrCreateUser(c.Request().Context(), user); err != nil {
		return err
	}

	c.Session().Put("user_id", user.ID)
	c.Session().Flash("success", "Signed in successfully via DigitalOcean")

	return c.Redirect("/")
}

// AuthDestroy logs out the currently logged in user.
func AuthDestroy(c seatbelt.Context) error {
	c.Session().Reset()
	return c.Redirect("/")
}
