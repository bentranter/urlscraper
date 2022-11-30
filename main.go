package main

import (
	"html/template"
	"os"
	"time"

	"github.com/bentranter/urlscraper/app/services/email"
	"github.com/bentranter/urlscraper/app/services/scraper"
	"github.com/bentranter/urlscraper/config"

	"github.com/bentranter/go-seatbelt"
	"github.com/dustin/go-humanize"
	"github.com/dustin/go-humanize/english"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	const (
		interval = 5 * time.Minute
		emailAt  = 14*time.Hour + 3*time.Minute
	)
	loc := time.Now().Local().Location() // TODO This won't work on most machines, you
	// should use a dedicated zone constant.

	// TODO Use the default high performance logger in production.
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().Timestamp().Logger()

	log.Debug().Msg("starting scraper, hit cmd-c to stop")
	log.Debug().Float64("interval", interval.Minutes()).
		Msg("running scraper with interval")
	go scraper.Scrape(interval)

	log.Debug().Msg("starting email update background job, hit cmd-c to stop")
	log.Debug().Str("start_time", emailAt.String()).
		Msg("running email update background job")
	go email.StartDailyUpdateJob(emailAt, loc, make(chan bool))

	app := seatbelt.New(seatbelt.Option{
		TemplateDir: "app/views",
		Reload:      true,
		Funcs: template.FuncMap{
			"timeago": humanize.Time,
			"pluralize": func(count int, word string) string {
				return english.PluralWord(count, word, "")
			},
		},
	})
	app = config.Routes(app)

	log.Fatal().Err(app.Start(":3000")).Msg("server error")
}
