package models

import (
	"context"
	"database/sql"
	"os"
	"os/user"
	"reflect"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

// db is global database connection.
var db *bun.DB

// DB returns a connection to the application database.
func DB() *bun.DB {
	return db
}

// init initializes a global connection to the database.
func init() {
	// Open a PostgreSQL database.
	u, err := user.Current()
	if err != nil {
		log.Error().Err(err).Msg("failed to get current user")
	}

	dsn := "postgres://" + u.Username + ":@localhost:5432/urlscraper?sslmode=disable"
	if s := os.Getenv("DSN"); s != "" {
		dsn = s
	}
	log.Info().Str("dsn", dsn).Msg("connecting to database")

	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

	// Create a Bun db on top of it.
	db = bun.NewDB(pgdb, pgdialect.New())

	// Print all queries to stdout.
	db.AddQueryHook(&zerologQueryHook{verbose: true})
}

// A zerologQueryHook satisfies Bun's QueryHook interface to provide support
// for structured query logs with zerolog.
//
// If verbose is set to true, it will log all queries. If it's set to false,
// only queries that error are logged.
type zerologQueryHook struct {
	verbose bool
}

func (z *zerologQueryHook) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx
}

func (z *zerologQueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if !z.verbose {
		switch event.Err {
		case nil, sql.ErrNoRows:
			return
		}
	}

	ll := log.Info()
	if event.Err != nil {
		ll = log.Error().
			Err(event.Err).
			Str("err_type", reflect.TypeOf(event.Err).String())
	}

	if v := ctx.Value("request_id"); v != nil {
		if requestID, ok := v.(string); ok {
			ll.Str("request_id", requestID)
		}
	}
	ll.Str("duration", time.Since(event.StartTime).String()).
		Str("operation", event.Operation()).
		Str("query", strings.ReplaceAll(event.Query, `"`, "")).
		Msg("query")
}
