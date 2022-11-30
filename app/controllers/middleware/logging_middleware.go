package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/bentranter/urlscraper/app/helpers"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Log logs all HTTP requests.
func Log(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		uuid := helpers.UUID()

		// Set the UUID as the request ID.
		r = r.WithContext(context.WithValue(r.Context(), "request_id", uuid))

		rs := stallHTTP(h, w, r)
		rs.sendResponse()

		duration := time.Since(now)

		// If the response takes longer than 300ms, warn as that's way too
		// long to spend handling a single request.
		var logEvent *zerolog.Event
		if duration > time.Millisecond*300 {
			logEvent = log.Warn()
		} else {
			logEvent = log.Info()
		}

		if rs.code >= 400 {
			logEvent = log.Error()
		}

		logEvent.Str("id", uuid).
			Str("duration", duration.String()).
			Str("path", r.URL.Path).
			Int("status", rs.code).
			Str("method", r.Method).
			Str("ip_addr", helpers.IPAddr(r)).
			Str("user_agent", r.UserAgent()).
			Msg("request")
	})
}
