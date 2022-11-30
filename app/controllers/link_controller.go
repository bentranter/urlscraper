package controllers

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"strconv"

	"github.com/bentranter/urlscraper/app/models"

	"github.com/bentranter/go-seatbelt"
	"github.com/rs/zerolog/log"
)

// LinkShow renders a saved link by ID.
func LinkShow(c seatbelt.Context) error {
	userID, ok := c.Session().Get("user_id").(int64)
	if !ok {
		return c.Redirect("/")
	}

	user, err := models.FindUser(c.Request().Context(), userID)
	if err != nil {
		return c.Redirect("/")
	}

	linkID, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return err
	}

	link, err := models.FindLink(c.Request().Context(), linkID)
	if err != nil {
		return err
	}

	return c.Render("links/show", map[string]interface{}{
		"User": user,
		"Link": link,
	})
}

// LinkNew renders the new link page.
func LinkNew(c seatbelt.Context) error {
	id, ok := c.Session().Get("user_id").(int64)
	if !ok {
		return c.Redirect("/")
	}

	user, err := models.FindUser(c.Request().Context(), id)
	if err != nil {
		return c.Redirect("/")
	}

	return c.Render("links/new", map[string]interface{}{"User": user})
}

// LinkCreate handles the POST request to create a new link.
func LinkCreate(c seatbelt.Context) error {
	id, ok := c.Session().Get("user_id").(int64)
	if !ok {
		return c.Redirect("/")
	}

	url := c.FormValue("url")
	name := c.FormValue("name")

	// The form shows the leading "https://" but doesn't submit it with the
	// response, so we need to add that manually.
	fullurl := "https://" + url

	resp, err := http.Get(fullurl)
	if err != nil {
		log.Error().Err(err).Str("url", fullurl).Msg("failed to GET url")
		return err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("failed to read body")
		return err
	}
	defer resp.Body.Close()

	h := sha256.New()
	h.Write(data)
	hash := base64.StdEncoding.EncodeToString(h.Sum(nil))

	link := &models.Link{
		URL:     fullurl,
		Name:    name,
		Hash:    hash,
		UserID:  id,
		Favicon: "https://s2.googleusercontent.com/s2/favicons?domain=" + url,
	}

	if _, err := models.CreateLink(c.Request().Context(), link); err != nil {
		return err
	}

	c.Session().Flash("success", "Link for "+link.Name+" saved successfully")

	return c.Redirect("/")
}

// LinkDestroy destroys a link.
func LinkDestroy(c seatbelt.Context) error {
	if _, ok := c.Session().Get("user_id").(int64); !ok {
		return c.Redirect("/")
	}

	id, err := strconv.ParseInt(c.PathParam("id"), 10, 64)
	if err != nil {
		return err
	}

	if err := models.DestroyLink(c.Request().Context(), id); err != nil {
		return err
	}

	c.Session().Flash("notice", "Delete successful")

	return c.Redirect("/")
}
