package mocks

import (
	"github.com/stretchr/testify/mock"
	"gopi.com/internal/app/chat"
	"gopi.com/internal/app/post"
	"gopi.com/internal/app/user"
	chatModel "gopi.com/internal/domain/chat/model"
	postModel "gopi.com/internal/domain/post/model"
	userModel "gopi.com/internal/domain/user/model"
)

// MockUserRepository implements the UserRepository interface for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *userModel.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id string) (*userModel.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*userModel.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(username string) (*userModel.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *userModel.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(limit, offset int) ([]*userModel.User, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmailAndPassword(email, password string) (*userModel.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetByOTP(email, otp string) (*userModel.User, error) {
	args := m.Called(email, otp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(id, newPassword string) error {
	args := m.Called(id, newPassword)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateOTP(id, otp string) error {
	args := m.Called(id, otp)
	return args.Error(0)
}

func (m *MockUserRepository) MarkAsVerified(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) GetAllUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetStaffUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetVerifiedUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserRepository) GetUnverifiedUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserRepository) EmailExists(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) UsernameExists(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

// MockEmailService implements the EmailService interface for testing
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendEmail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordResetEmail(to, resetToken string) error {
	args := m.Called(to, resetToken)
	return args.Error(0)
}

func (m *MockEmailService) SendWelcomeEmail(to, username string) error {
	args := m.Called(to, username)
	return args.Error(0)
}

// MockGroupRepository implements the GroupRepository interface for testing
type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) Create(group *chatModel.Group) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockGroupRepository) GetByID(id string) (*chatModel.Group, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chatModel.Group), args.Error(1)
}

func (m *MockGroupRepository) GetBySlug(slug string) (*chatModel.Group, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chatModel.Group), args.Error(1)
}

func (m *MockGroupRepository) GetByCreatorID(creatorID string) ([]*chatModel.Group, error) {
	args := m.Called(creatorID)
	return args.Get(0).([]*chatModel.Group), args.Error(1)
}

func (m *MockGroupRepository) GetByMemberID(memberID string) ([]*chatModel.Group, error) {
	args := m.Called(memberID)
	return args.Get(0).([]*chatModel.Group), args.Error(1)
}

func (m *MockGroupRepository) Update(group *chatModel.Group) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockGroupRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockGroupRepository) AddMember(groupID, memberID string) error {
	args := m.Called(groupID, memberID)
	return args.Error(0)
}

func (m *MockGroupRepository) RemoveMember(groupID, memberID string) error {
	args := m.Called(groupID, memberID)
	return args.Error(0)
}

func (m *MockGroupRepository) SearchByName(query string, limit, offset int) ([]*chatModel.Group, error) {
	args := m.Called(query, limit, offset)
	return args.Get(0).([]*chatModel.Group), args.Error(1)
}

// MockMessageRepository implements the MessageRepository interface for testing
type MockMessageRepository struct {
	mock.Mock
}

func (m *MockMessageRepository) Create(message *chatModel.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockMessageRepository) GetByID(id string) (*chatModel.Message, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chatModel.Message), args.Error(1)
}

func (m *MockMessageRepository) GetByGroupID(groupID string, limit, offset int) ([]*chatModel.Message, error) {
	args := m.Called(groupID, limit, offset)
	return args.Get(0).([]*chatModel.Message), args.Error(1)
}

func (m *MockMessageRepository) GetBySenderID(senderID string) ([]*chatModel.Message, error) {
	args := m.Called(senderID)
	return args.Get(0).([]*chatModel.Message), args.Error(1)
}

func (m *MockMessageRepository) Update(message *chatModel.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockMessageRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

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

// MockChatService implements the ChatService interface for testing
type MockChatService struct {
	mock.Mock
	*chat.ChatService // Embed concrete type for compatibility
}

func (m *MockChatService) CreateGroup(creatorID, name, image string, memberIDs []string) (*chatModel.Group, error) {
	args := m.Called(creatorID, name, image, memberIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chatModel.Group), args.Error(1)
}

func (m *MockChatService) GetGroupByID(id string) (*chatModel.Group, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chatModel.Group), args.Error(1)
}

func (m *MockChatService) GetGroupBySlug(slug string) (*chatModel.Group, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chatModel.Group), args.Error(1)
}

func (m *MockChatService) GetGroupsByMember(memberID string) ([]*chatModel.Group, error) {
	args := m.Called(memberID)
	return args.Get(0).([]*chatModel.Group), args.Error(1)
}

func (m *MockChatService) UpdateGroup(group *chatModel.Group) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockChatService) DeleteGroup(groupID, requesterID string) error {
	args := m.Called(groupID, requesterID)
	return args.Error(0)
}

func (m *MockChatService) AddMemberToGroup(groupID, memberID, requesterID string) error {
	args := m.Called(groupID, memberID, requesterID)
	return args.Error(0)
}

func (m *MockChatService) RemoveMemberFromGroup(groupID, memberID, requesterID string) error {
	args := m.Called(groupID, memberID, requesterID)
	return args.Error(0)
}

func (m *MockChatService) SendMessage(senderID, groupID, content string) (*chatModel.Message, error) {
	args := m.Called(senderID, groupID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*chatModel.Message), args.Error(1)
}

func (m *MockChatService) GetMessagesByGroup(groupID string, limit, offset int) ([]*chatModel.Message, error) {
	args := m.Called(groupID, limit, offset)
	return args.Get(0).([]*chatModel.Message), args.Error(1)
}

func (m *MockChatService) SearchGroupsByName(query string, limit, offset int) ([]*chatModel.Group, error) {
	args := m.Called(query, limit, offset)
	return args.Get(0).([]*chatModel.Group), args.Error(1)
}

// MockUserService implements the UserService interface for testing
type MockUserService struct {
	mock.Mock
	*user.UserService // Embed concrete type for compatibility
}

func (m *MockUserService) RegisterUser(username, email, firstName, lastName, password string, height, weight float64) (*userModel.User, error) {
	args := m.Called(username, email, firstName, lastName, password, height, weight)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserService) LoginUser(email, password string) (*userModel.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id string) (*userModel.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(email string) (*userModel.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserService) GetUserByUsername(username string) (*userModel.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(user *userModel.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) ChangePassword(id, oldPassword, newPassword string) error {
	args := m.Called(id, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockUserService) ResetPassword(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockUserService) VerifyOTP(email, otp string) (*userModel.User, error) {
	args := m.Called(email, otp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userModel.User), args.Error(1)
}

func (m *MockUserService) SendPasswordResetEmail(email string) error {
	args := m.Called(email)
	return args.Error(0)
}

func (m *MockUserService) GetAllUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserService) GetStaffUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserService) GetVerifiedUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
}

func (m *MockUserService) GetUnverifiedUsers() ([]*userModel.User, error) {
	args := m.Called()
	return args.Get(0).([]*userModel.User), args.Error(1)
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
