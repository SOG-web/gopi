package repo

import postModel "gopi.com/internal/domain/post/model"

// PostRepository abstracts persistence for posts (articles)
type PostRepository interface {
	Create(post *postModel.Post) error
	Update(post *postModel.Post) error
	Delete(id string) error
	GetByID(id string) (*postModel.Post, error)
	GetBySlug(slug string) (*postModel.Post, error)
	ListPublished(limit, offset int) ([]*postModel.Post, error)
	ListByAuthor(authorID string, limit, offset int) ([]*postModel.Post, error)
	SearchPublished(query string, limit, offset int) ([]*postModel.Post, error)
}

// CommentRepository abstracts persistence for comments
type CommentRepository interface {
	Create(comment *postModel.Comment) error
	Update(comment *postModel.Comment) error
	Delete(id string) error
	GetByID(id string) (*postModel.Comment, error)
	ListByTarget(targetType, targetID string, limit, offset int) ([]*postModel.Comment, error)
	ListByAuthor(authorID string, limit, offset int) ([]*postModel.Comment, error)
	ListReplies(parentID string, limit, offset int) ([]*postModel.Comment, error)
}
