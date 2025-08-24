package post

import (
	"errors"
	"time"
	"unicode/utf8"

	postModel "gopi.com/internal/domain/post/model"
	postRepo "gopi.com/internal/domain/post/repo"
	"gopi.com/internal/domain/model"
	"gopi.com/internal/lib/id"
)

type Service struct {
	postRepo    postRepo.PostRepository
	commentRepo postRepo.CommentRepository
}

func NewPostService(postRepo postRepo.PostRepository, commentRepo postRepo.CommentRepository) *Service {
	return &Service{postRepo: postRepo, commentRepo: commentRepo}
}

// Posts

func (s *Service) CreatePost(authorID, title, content, coverImageURL string, publish bool) (*postModel.Post, error) {
	if title == "" || utf8.RuneCountInString(title) > 200 {
		return nil, errors.New("title must be 1-200 characters")
	}
	if utf8.RuneCountInString(content) == 0 {
		return nil, errors.New("content is required")
	}

	now := time.Now()
	p := &postModel.Post{
		Base: model.Base{ID: id.New(), CreatedAt: now, UpdatedAt: now},
		Title:         title,
		Slug:          generateSlug(title),
		Content:       content,
		AuthorID:      authorID,
		CoverImageURL: coverImageURL,
		IsPublished:   publish,
	}
	if publish {
		p.PublishedAt = &now
	}
	if err := s.postRepo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) UpdatePost(postID, title, content, coverImageURL string) (*postModel.Post, error) {
	p, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, err
	}
	if title != "" {
		if utf8.RuneCountInString(title) > 200 {
			return nil, errors.New("title must be at most 200 characters")
		}
		p.Title = title
		p.Slug = generateSlug(title)
	}
	if content != "" {
		p.Content = content
	}
	if coverImageURL != "" {
		p.CoverImageURL = coverImageURL
	}
	p.UpdatedAt = time.Now()
	if err := s.postRepo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) PublishPost(postID string) (*postModel.Post, error) {
	p, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	p.IsPublished = true
	p.PublishedAt = &now
	p.UpdatedAt = now
	if err := s.postRepo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) UnpublishPost(postID string) (*postModel.Post, error) {
	p, err := s.postRepo.GetByID(postID)
	if err != nil {
		return nil, err
	}
	p.IsPublished = false
	p.PublishedAt = nil
	p.UpdatedAt = time.Now()
	if err := s.postRepo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) DeletePost(postID string) error {
	return s.postRepo.Delete(postID)
}

func (s *Service) GetPostByID(id string) (*postModel.Post, error) {
	return s.postRepo.GetByID(id)
}

func (s *Service) GetPostBySlug(slug string) (*postModel.Post, error) {
	return s.postRepo.GetBySlug(slug)
}

func (s *Service) ListPublished(limit, offset int) ([]*postModel.Post, error) {
	return s.postRepo.ListPublished(limit, offset)
}

func (s *Service) ListByAuthor(authorID string, limit, offset int) ([]*postModel.Post, error) {
	return s.postRepo.ListByAuthor(authorID, limit, offset)
}

func (s *Service) SearchPublished(query string, limit, offset int) ([]*postModel.Post, error) {
	return s.postRepo.SearchPublished(query, limit, offset)
}

// Comments

func (s *Service) CreateComment(authorID, targetType, targetID, content string, parentID *string) (*postModel.Comment, error) {
	if utf8.RuneCountInString(content) == 0 || utf8.RuneCountInString(content) > 2000 {
		return nil, errors.New("content must be 1-2000 characters")
	}
	if targetType == "" || targetID == "" {
		return nil, errors.New("target_type and target_id are required")
	}
	now := time.Now()
	c := &postModel.Comment{
		Base:       model.Base{ID: id.New(), CreatedAt: now, UpdatedAt: now},
		AuthorID:   authorID,
		Content:    content,
		TargetType: targetType,
		TargetID:   targetID,
		ParentID:   parentID,
		IsDeleted:  false,
	}
	if err := s.commentRepo.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) UpdateComment(commentID, authorID, content string) (*postModel.Comment, error) {
	if utf8.RuneCountInString(content) == 0 || utf8.RuneCountInString(content) > 2000 {
		return nil, errors.New("content must be 1-2000 characters")
	}
	c, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return nil, err
	}
	if c.AuthorID != authorID {
		return nil, errors.New("only the author can update the comment")
	}
	c.Content = content
	c.UpdatedAt = time.Now()
	if err := s.commentRepo.Update(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) DeleteComment(commentID, requesterID string) error {
	c, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		return err
	}
	if c.AuthorID != requesterID {
		return errors.New("only the author can delete the comment")
	}
	return s.commentRepo.Delete(commentID)
}

func (s *Service) ListCommentsByTarget(targetType, targetID string, limit, offset int) ([]*postModel.Comment, error) {
	return s.commentRepo.ListByTarget(targetType, targetID, limit, offset)
}

func (s *Service) ListReplies(parentID string, limit, offset int) ([]*postModel.Comment, error) {
	return s.commentRepo.ListReplies(parentID, limit, offset)
}

// helper
func generateSlug(title string) string {
	return title + "-" + id.New()[:8]
}
