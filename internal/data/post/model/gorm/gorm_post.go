package gorm

import (
	"time"

	postModel "gopi.com/internal/domain/post/model"
	"gopi.com/internal/domain/model"
	"gopi.com/internal/lib/id"
	"gorm.io/gorm"
)

// GORM models

type Post struct {
	ID            string     `gorm:"type:varchar(255);primary_key"`
	Title         string     `gorm:"not null"`
	Slug          string     `gorm:"uniqueIndex;size:255;not null"`
	Content       string     `gorm:"type:text;not null"`
	AuthorID      string     `gorm:"index;not null"`
	CoverImageURL string
	IsPublished   bool       `gorm:"index"`
	PublishedAt   *time.Time `gorm:"index"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Comment struct {
	ID         string     `gorm:"type:varchar(255);primary_key"`
	AuthorID   string     `gorm:"index;not null"`
	Content    string     `gorm:"type:text;not null"`
	TargetType string     `gorm:"index:idx_target,priority:1;size:50;not null"`
	TargetID   string     `gorm:"index:idx_target,priority:2;size:255;not null"`
	ParentID   *string    `gorm:"index"`
	IsDeleted  bool       `gorm:"index"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (p *Post) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = id.New()
	}
	return
}

func (c *Comment) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID = id.New()
	}
	return
}

// Converters

func FromDomainPost(p *postModel.Post) *Post {
	return &Post{
		ID:            p.ID,
		Title:         p.Title,
		Slug:          p.Slug,
		Content:       p.Content,
		AuthorID:      p.AuthorID,
		CoverImageURL: p.CoverImageURL,
		IsPublished:   p.IsPublished,
		PublishedAt:   p.PublishedAt,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func ToDomainPost(p *Post) *postModel.Post {
	return &postModel.Post{
		Base: model.Base{
			ID:        p.ID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		},
		Title:         p.Title,
		Slug:          p.Slug,
		Content:       p.Content,
		AuthorID:      p.AuthorID,
		CoverImageURL: p.CoverImageURL,
		IsPublished:   p.IsPublished,
		PublishedAt:   p.PublishedAt,
	}
}

func FromDomainComment(c *postModel.Comment) *Comment {
	return &Comment{
		ID:         c.ID,
		AuthorID:   c.AuthorID,
		Content:    c.Content,
		TargetType: c.TargetType,
		TargetID:   c.TargetID,
		ParentID:   c.ParentID,
		IsDeleted:  c.IsDeleted,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

func ToDomainComment(c *Comment) *postModel.Comment {
	return &postModel.Comment{
		Base: model.Base{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		},
		AuthorID:   c.AuthorID,
		Content:    c.Content,
		TargetType: c.TargetType,
		TargetID:   c.TargetID,
		ParentID:   c.ParentID,
		IsDeleted:  c.IsDeleted,
	}
}
