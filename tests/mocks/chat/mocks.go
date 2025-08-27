package chat

import (
	"github.com/stretchr/testify/mock"
	"gopi.com/internal/app/chat"
	chatModel "gopi.com/internal/domain/chat/model"
)

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
