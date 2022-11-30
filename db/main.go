package main

import (
	"context"
	"os"

	"github.com/bentranter/urlscraper/app/models"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun/migrate"
)

func main() {
	// TODO Use the default high performance logger in production.
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().Timestamp().Logger()

	migrations := migrate.NewMigrations()
	if err := migrations.DiscoverCaller(); err != nil {
		log.Fatal().Err(err).Msg("failed to discover callers")
	}

	migrator := migrate.NewMigrator(models.DB(), migrations)

	if err := migrator.Init(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("failed to init database tables for migration")
	}

	var group *migrate.MigrationGroup
	var err error

	if len(os.Args) > 1 && os.Args[1] == "down" {
		group, err = migrator.Rollback(context.Background())
	} else {
		group, err = migrator.Migrate(context.Background())
	}
	if err != nil {
		log.Fatal().Err(err).Msg("failed to migrate database")
	}

	if group.ID == 0 {
		log.Info().Msg("there are no new migrations to run")
		return
	}

	log.Info().Str("group", group.String()).Msg("migrated database")
}
