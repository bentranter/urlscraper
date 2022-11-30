package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

// A UserRelation is used to optionally include data from other tables
// associated with a user in select queries.
type UserRelation string

const (
	// UserRelationLinks is used to include users' links.
	UserRelationLinks UserRelation = "Links"
)

var (
	ErrMissingEmail = errors.New("email is required")
)

// A User is a user.
type User struct {
	ID        int64     `json:"-"`
	Email     string    `json:"email"`
	FirstName string    `json:"given_name"`
	LastName  string    `json:"family_name"`
	Picture   string    `json:"picture"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Links     []*Link   `bun:"rel:has-many"`
}

// FindUser finds the user with the given ID.
func FindUser(ctx context.Context, id int64, relations ...UserRelation) (*User, error) {
	user := &User{ID: id}

	q := db.NewSelect().
		Model(user).
		WherePK()

	for _, r := range relations {
		q = q.Relation(string(r), func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Order("last_change_at DESC")
		})
	}

	err := q.Scan(ctx)
	return user, err
}

// FindOrCreateUser finds a user by email if they exist, and creates the user
// otherwise.
func FindOrCreateUser(ctx context.Context, user *User) (*User, error) {
	if user.Email == "" {
		return nil, ErrMissingEmail
	}

	err := db.NewSelect().
		Model(user).
		Where("email = ?", user.Email).
		Scan(ctx)
	if err == nil {
		return user, nil
	}
	if err != nil && err != sql.ErrNoRows {
		log.Error().Err(err).Msg("failed to select user")
		return nil, err
	}

	// Something is up with the column scanning logic when inserting empty
	// timestamps, so we'll manually set them for inserts.
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err = db.NewInsert().
		Model(user).
		Returning("*").
		Exec(ctx)
	return user, err
}
