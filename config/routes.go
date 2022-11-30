package config

import (
	"github.com/bentranter/urlscraper/app/controllers"
	"github.com/bentranter/urlscraper/app/controllers/middleware"

	"github.com/bentranter/go-seatbelt"
)

// Routes registers HTTP controllers as handlers on a Seatbelt application.
func Routes(app *seatbelt.App) *seatbelt.App {
	app.UseStd(middleware.Log)

	app.Get("/", controllers.HomeIndex)

	app.Get("/account/login", controllers.AuthIndex)

	app.Get("/auth/digitalocean/authorize", controllers.AuthRedirect)
	app.Get("/auth/digitalocean/callback", controllers.AuthCallback)
	app.Post("/auth/destroy", controllers.AuthDestroy)

	app.Get("/links/{id}", controllers.LinkShow)
	app.Post("/links", controllers.LinkCreate)
	app.Get("/links/new", controllers.LinkNew)
	app.Post("/links/{id}/destroy", controllers.LinkDestroy)

	app.FileServer("/public", "public")

	return app
}
