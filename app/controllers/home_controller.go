package controllers

import (
	"github.com/bentranter/urlscraper/app/models"

	"github.com/bentranter/go-seatbelt"
	"github.com/rs/zerolog/log"
)

// HomeIndex renders the home page.
func HomeIndex(c seatbelt.Context) error {
	data := make(map[string]interface{})

	id, ok := c.Session().Get("user_id").(int64)
	if ok {
		user, err := models.FindUser(c.Request().Context(), id, models.UserRelationLinks)
		if err != nil {
			log.Debug().Err(err).Int64("id", id).Msg("failed to find user by id")
		} else {
			data["User"] = user
		}
	}

	return c.Render("home/index", data)
}
