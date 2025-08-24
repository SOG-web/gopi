package model

import (
	"time"

	"gopi.com/internal/domain/model"
)

// Post represents an article authored by a staff/admin user
// It is publicly readable when published.
type Post struct {
	model.Base
	Title          string     `json:"title"`
	Slug           string     `json:"slug"`
	Content        string     `json:"content"`
	AuthorID       string     `json:"author_id"`
	CoverImageURL  string     `json:"cover_image_url"`
	IsPublished    bool       `json:"is_published"`
	PublishedAt    *time.Time `json:"published_at"`
}
