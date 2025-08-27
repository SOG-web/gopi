package post

import (
	"github.com/stretchr/testify/mock"
	"gopi.com/internal/app/post"
	postModel "gopi.com/internal/domain/post/model"
)

// MockPostRepository implements the PostRepository interface for testing
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(post *postModel.Post) error {
	args := m.Called(post)
	return args.Error(0)
}

func (m *MockPostRepository) Update(post *postModel.Post) error {
	args := m.Called(post)
	return args.Error(0)
}

func (m *MockPostRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPostRepository) GetByID(id string) (*postModel.Post, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postModel.Post), args.Error(1)
}

func (m *MockPostRepository) GetBySlug(slug string) (*postModel.Post, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postModel.Post), args.Error(1)
}

func (m *MockPostRepository) ListPublished(limit, offset int) ([]*postModel.Post, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]*postModel.Post), args.Error(1)
}

func (m *MockPostRepository) ListByAuthor(authorID string, limit, offset int) ([]*postModel.Post, error) {
	args := m.Called(authorID, limit, offset)
	return args.Get(0).([]*postModel.Post), args.Error(1)
}

func (m *MockPostRepository) SearchPublished(query string, limit, offset int) ([]*postModel.Post, error) {
	args := m.Called(query, limit, offset)
	return args.Get(0).([]*postModel.Post), args.Error(1)
}

// MockCommentRepository implements the CommentRepository interface for testing
type MockCommentRepository struct {
	mock.Mock
}

func (m *MockCommentRepository) Create(comment *postModel.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepository) Update(comment *postModel.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *MockCommentRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCommentRepository) GetByID(id string) (*postModel.Comment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postModel.Comment), args.Error(1)
}

func (m *MockCommentRepository) ListByTarget(targetType, targetID string, limit, offset int) ([]*postModel.Comment, error) {
	args := m.Called(targetType, targetID, limit, offset)
	return args.Get(0).([]*postModel.Comment), args.Error(1)
}

func (m *MockCommentRepository) ListByAuthor(authorID string, limit, offset int) ([]*postModel.Comment, error) {
	args := m.Called(authorID, limit, offset)
	return args.Get(0).([]*postModel.Comment), args.Error(1)
}

func (m *MockCommentRepository) ListReplies(parentID string, limit, offset int) ([]*postModel.Comment, error) {
	args := m.Called(parentID, limit, offset)
	return args.Get(0).([]*postModel.Comment), args.Error(1)
}

// MockPostService implements the PostService interface for testing
type MockPostService struct {
	mock.Mock
	*post.Service // Embed concrete type for compatibility
}

func (m *MockPostService) CreatePost(authorID, title, content, coverImageURL string, publish bool) (*postModel.Post, error) {
	args := m.Called(authorID, title, content, coverImageURL, publish)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postModel.Post), args.Error(1)
}

func (m *MockPostService) UpdatePost(id, title, content, coverImageURL string, publish bool) (*postModel.Post, error) {
	args := m.Called(id, title, content, coverImageURL, publish)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postModel.Post), args.Error(1)
}

func (m *MockPostService) DeletePost(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPostService) GetPostByID(id string) (*postModel.Post, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postModel.Post), args.Error(1)
}

func (m *MockPostService) GetPostBySlug(slug string) (*postModel.Post, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postModel.Post), args.Error(1)
}

func (m *MockPostService) ListPublishedPosts(limit, offset int) ([]*postModel.Post, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]*postModel.Post), args.Error(1)
}

func (m *MockPostService) ListPostsByAuthor(authorID string, limit, offset int) ([]*postModel.Post, error) {
	args := m.Called(authorID, limit, offset)
	return args.Get(0).([]*postModel.Post), args.Error(1)
}

func (m *MockPostService) SearchPublishedPosts(query string, limit, offset int) ([]*postModel.Post, error) {
	args := m.Called(query, limit, offset)
	return args.Get(0).([]*postModel.Post), args.Error(1)
}

func (m *MockPostService) CreateComment(authorID, content, targetType, targetID string, parentID *string) (*postModel.Comment, error) {
	args := m.Called(authorID, content, targetType, targetID, parentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postModel.Comment), args.Error(1)
}

func (m *MockPostService) UpdateComment(id, content string) (*postModel.Comment, error) {
	args := m.Called(id, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postModel.Comment), args.Error(1)
}

func (m *MockPostService) DeleteComment(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPostService) GetCommentByID(id string) (*postModel.Comment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*postModel.Comment), args.Error(1)
}

func (m *MockPostService) ListCommentsByTarget(targetType, targetID string, limit, offset int) ([]*postModel.Comment, error) {
	args := m.Called(targetType, targetID, limit, offset)
	return args.Get(0).([]*postModel.Comment), args.Error(1)
}

func (m *MockPostService) ListCommentsByAuthor(authorID string, limit, offset int) ([]*postModel.Comment, error) {
	args := m.Called(authorID, limit, offset)
	return args.Get(0).([]*postModel.Comment), args.Error(1)
}

func (m *MockPostService) ListCommentReplies(parentID string, limit, offset int) ([]*postModel.Comment, error) {
	args := m.Called(parentID, limit, offset)
	return args.Get(0).([]*postModel.Comment), args.Error(1)
}
