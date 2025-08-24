package repo

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	gormmodel "gopi.com/internal/data/post/model/gorm"
	postModel "gopi.com/internal/domain/post/model"
	postRepo "gopi.com/internal/domain/post/repo"
)

// Post repository implementation

type GormPostRepository struct {
	db *gorm.DB
}

func NewGormPostRepository(db *gorm.DB) postRepo.PostRepository {
	return &GormPostRepository{db: db}
}

func (r *GormPostRepository) Create(post *postModel.Post) error {
	dbPost := gormmodel.FromDomainPost(post)
	if err := r.db.Create(&dbPost).Error; err != nil {
		return err
	}
	*post = *gormmodel.ToDomainPost(dbPost)
	return nil
}

func (r *GormPostRepository) Update(post *postModel.Post) error {
	dbPost := gormmodel.FromDomainPost(post)
	if err := r.db.Save(&dbPost).Error; err != nil {
		return err
	}
	*post = *gormmodel.ToDomainPost(dbPost)
	return nil
}

func (r *GormPostRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.Post{}, "id = ?", id).Error
}

func (r *GormPostRepository) GetByID(id string) (*postModel.Post, error) {
	var p gormmodel.Post
	if err := r.db.First(&p, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return gormmodel.ToDomainPost(&p), nil
}

func (r *GormPostRepository) GetBySlug(slug string) (*postModel.Post, error) {
	var p gormmodel.Post
	if err := r.db.Where("slug = ?", slug).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return gormmodel.ToDomainPost(&p), nil
}

func (r *GormPostRepository) ListPublished(limit, offset int) ([]*postModel.Post, error) {
	var posts []gormmodel.Post
	if err := r.db.Where("is_published = ?", true).
		Order("published_at DESC, created_at DESC").
		Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		return nil, err
	}
	result := make([]*postModel.Post, 0, len(posts))
	for i := range posts {
		result = append(result, gormmodel.ToDomainPost(&posts[i]))
	}
	return result, nil
}

func (r *GormPostRepository) ListByAuthor(authorID string, limit, offset int) ([]*postModel.Post, error) {
	var posts []gormmodel.Post
	if err := r.db.Where("author_id = ?", authorID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		return nil, err
	}
	result := make([]*postModel.Post, 0, len(posts))
	for i := range posts {
		result = append(result, gormmodel.ToDomainPost(&posts[i]))
	}
	return result, nil
}

func (r *GormPostRepository) SearchPublished(query string, limit, offset int) ([]*postModel.Post, error) {
	var posts []gormmodel.Post
	q := "%" + strings.ToLower(query) + "%"
	if err := r.db.Where("is_published = ? AND (LOWER(title) LIKE ? OR LOWER(content) LIKE ?)", true, q, q).
		Order("published_at DESC, created_at DESC").
		Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		return nil, err
	}
	result := make([]*postModel.Post, 0, len(posts))
	for i := range posts {
		result = append(result, gormmodel.ToDomainPost(&posts[i]))
	}
	return result, nil
}

// Comment repository implementation

type GormCommentRepository struct {
	db *gorm.DB
}

func NewGormCommentRepository(db *gorm.DB) postRepo.CommentRepository {
	return &GormCommentRepository{db: db}
}

func (r *GormCommentRepository) Create(comment *postModel.Comment) error {
	dbComment := gormmodel.FromDomainComment(comment)
	if err := r.db.Create(&dbComment).Error; err != nil {
		return err
	}
	*comment = *gormmodel.ToDomainComment(dbComment)
	return nil
}

func (r *GormCommentRepository) Update(comment *postModel.Comment) error {
	dbComment := gormmodel.FromDomainComment(comment)
	if err := r.db.Save(&dbComment).Error; err != nil {
		return err
	}
	*comment = *gormmodel.ToDomainComment(dbComment)
	return nil
}

func (r *GormCommentRepository) Delete(id string) error {
	return r.db.Delete(&gormmodel.Comment{}, "id = ?", id).Error
}

func (r *GormCommentRepository) GetByID(id string) (*postModel.Comment, error) {
	var c gormmodel.Comment
	if err := r.db.First(&c, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return gormmodel.ToDomainComment(&c), nil
}

func (r *GormCommentRepository) ListByTarget(targetType, targetID string, limit, offset int) ([]*postModel.Comment, error) {
	var comments []gormmodel.Comment
	if err := r.db.Where("target_type = ? AND target_id = ? AND is_deleted = ?", targetType, targetID, false).
		Order("created_at ASC").
		Limit(limit).Offset(offset).Find(&comments).Error; err != nil {
		return nil, err
	}
	result := make([]*postModel.Comment, 0, len(comments))
	for i := range comments {
		result = append(result, gormmodel.ToDomainComment(&comments[i]))
	}
	return result, nil
}

func (r *GormCommentRepository) ListByAuthor(authorID string, limit, offset int) ([]*postModel.Comment, error) {
	var comments []gormmodel.Comment
	if err := r.db.Where("author_id = ? AND is_deleted = ?", authorID, false).
		Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&comments).Error; err != nil {
		return nil, err
	}
	result := make([]*postModel.Comment, 0, len(comments))
	for i := range comments {
		result = append(result, gormmodel.ToDomainComment(&comments[i]))
	}
	return result, nil
}

func (r *GormCommentRepository) ListReplies(parentID string, limit, offset int) ([]*postModel.Comment, error) {
	var comments []gormmodel.Comment
	if err := r.db.Where("parent_id = ? AND is_deleted = ?", parentID, false).
		Order("created_at ASC").
		Limit(limit).Offset(offset).Find(&comments).Error; err != nil {
		return nil, err
	}
	result := make([]*postModel.Comment, 0, len(comments))
	for i := range comments {
		result = append(result, gormmodel.ToDomainComment(&comments[i]))
	}
	return result, nil
}
