package email

import (
	"context"
	"fmt"
	"time"

	"github.com/bentranter/urlscraper/app/models"
	"github.com/rs/zerolog/log"
)

// StartDailyUpdateJob starts a process that sends email updates each day at
// the given time of day.
//
// StartDailyUpdateJob is a blocking function, so you may want to call it in
// a goroutine.
func StartDailyUpdateJob(at time.Duration, loc *time.Location, quit chan bool) {
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, int(at.Nanoseconds()), loc)

	log.Debug().Str("at", start.String()).Msg("initial start time")

	// If the start time is after the current time, schedule the inital start
	// for tomorrow instead.
	if now.After(start) {
		start = start.AddDate(0, 0, 1)
	}

	log.Debug().Str("at", start.String()).Msg("final start time")

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-quit:
			fmt.Println("Done!")
			return

		case t := <-ticker.C:
			log.Debug().Msgf("ticked at %s", t.String())

			if t.After(start) {
				links, err := models.AllLinks(context.Background())
				if err != nil {
					log.Error().Err(err).Msg("failed to get all links")
					continue
				}

				for _, link := range links {
					if isWithinLast24Hours(link.UpdatedAt) {
						Send("test@example.com", "Test User", "Updates for current day", link.Name+" changed!")
					}
				}

				// Advance the next start date to tomorrow.
				start = start.AddDate(0, 0, 1)
			} else {
				log.Debug().
					Str("start", start.String()).
					Str("tick", t.String()).
					Msg("not firing")
			}
		}
	}
}

// isWithinLast24Hours checks if time t is within the last 24 hours.
func isWithinLast24Hours(t time.Time) bool {
	end := time.Now().Local().Add(1 * time.Nanosecond)
	start := time.Date(end.Year(), end.Month(), end.Day()-1, 0, 0, 0, end.Nanosecond(), end.Location())

	if start.Before(end) {
		return !t.Before(start) && !t.After(end)
	}
	if start.Equal(end) {
		return t.Equal(start)
	}
	return !start.After(t) || !end.Before(t)
}
