package models

import (
	"context"
	"errors"
	"time"
)

var (
	ErrLinkMissingUserID = errors.New("link is missing a user id")
	ErrLinkMissingURL    = errors.New("link is missing a url")
)

// A Link stores a link and its hash in the database.
//
// A link belongs to a user.
type Link struct {
	ID           int64
	Name         string
	URL          string
	Hash         string
	Favicon      string
	Screenshot   string
	ScrapeCount  int
	LastChangeAt time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	UserID       int64
}

// FindLink selects a link by ID.
func FindLink(ctx context.Context, id int64) (*Link, error) {
	link := &Link{ID: id}

	err := db.NewSelect().
		Model(link).
		WherePK().
		Scan(ctx)

	return link, err
}

// AllLinks selects all links.
func AllLinks(ctx context.Context) ([]*Link, error) {
	links := make([]*Link, 0)

	err := db.NewSelect().
		Model(&links).
		Scan(ctx)

	return links, err
}

// CreateLink creates a new link. A user ID must be present on the link.
func CreateLink(ctx context.Context, link *Link) (*Link, error) {
	if link.UserID == 0 {
		return nil, ErrLinkMissingUserID
	}
	if link.URL == "" {
		return nil, ErrLinkMissingURL
	}

	// Default the name to the URL if a name isn't provided.
	if link.Name == "" {
		link.Name = link.URL
	}

	now := time.Now()
	link.LastChangeAt = now
	link.CreatedAt = now
	link.UpdatedAt = now

	_, err := db.NewInsert().
		Model(link).
		Returning("*").
		Exec(ctx)
	return link, err
}

// Update updates the link with the given ID.
func UpdateLink(ctx context.Context, id int64, link *Link) (*Link, error) {
	link.ID = id

	_, err := db.NewUpdate().
		Model(link).
		WherePK().
		Exec(ctx)

	return link, err
}

// DestroyLink deletes the link with the given ID.
func DestroyLink(ctx context.Context, id int64) error {
	_, err := db.NewDelete().
		Model(&Link{ID: id}).
		WherePK().
		Exec(ctx)

	return err
}
