package chat_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopi.com/internal/app/chat"
	chatModel "gopi.com/internal/domain/chat/model"
	"gopi.com/internal/domain/model"
	chatMocks "gopi.com/tests/mocks/chat"
)

func TestChatService_CreateGroup(t *testing.T) {
	tests := []struct {
		name        string
		creatorID   string
		groupName   string
		image       string
		memberIDs   []string
		expectedErr error
		mockSetup   func(*chatMocks.MockGroupRepository)
	}{
		{
			name:      "successful group creation",
			creatorID: "user123",
			groupName: "Test Group",
			image:     "test.jpg",
			memberIDs: []string{"member1", "member2"},
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				mockRepo.On("Create", mock.MatchedBy(func(g *chatModel.Group) bool {
					return g.Name == "Test Group" &&
						g.CreatorID == "user123" &&
						len(g.MemberIDs) == 3 && // creator + 2 members
						g.Image == "test.jpg"
				})).Return(nil)
			},
		},
		{
			name:        "group name too long",
			creatorID:   "user123",
			groupName:   "This is a very long group name that exceeds the 20 character limit",
			image:       "test.jpg",
			memberIDs:   []string{"member1"},
			expectedErr: errors.New("group name must be at most 20 characters"),
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				// No repository call expected due to validation failure
			},
		},
		{
			name:        "repository error",
			creatorID:   "user123",
			groupName:   "Test Group",
			image:       "test.jpg",
			memberIDs:   []string{"member1"},
			expectedErr: errors.New("repository error"),
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				mockRepo.On("Create", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupRepo := new(chatMocks.MockGroupRepository)
			mockMessageRepo := new(chatMocks.MockMessageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockGroupRepo)
			}

			chatService := chat.NewChatService(mockGroupRepo, mockMessageRepo)

			group, err := chatService.CreateGroup(tt.creatorID, tt.groupName, tt.image, tt.memberIDs)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, group)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, group)
				assert.Equal(t, tt.groupName, group.Name)
				assert.Equal(t, tt.creatorID, group.CreatorID)
				assert.Equal(t, tt.image, group.Image)
				assert.Contains(t, group.MemberIDs, tt.creatorID) // Creator should be in members
			}

			mockGroupRepo.AssertExpectations(t)
		})
	}
}

func TestChatService_GetGroupByID(t *testing.T) {
	tests := []struct {
		name        string
		groupID     string
		expectedErr error
		mockSetup   func(*chatMocks.MockGroupRepository)
	}{
		{
			name:    "successful retrieval",
			groupID: "group123",
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				expectedGroup := &chatModel.Group{
					Base: model.Base{
						ID:        "group123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "user123",
					MemberIDs: []string{"user123", "member1"},
					Slug:      "test-group-abc123",
				}
				mockRepo.On("GetByID", "group123").Return(expectedGroup, nil)
			},
		},
		{
			name:        "group not found",
			groupID:     "nonexistent",
			expectedErr: errors.New("group not found"),
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				mockRepo.On("GetByID", "nonexistent").Return(nil, errors.New("group not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupRepo := new(chatMocks.MockGroupRepository)
			mockMessageRepo := new(chatMocks.MockMessageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockGroupRepo)
			}

			chatService := chat.NewChatService(mockGroupRepo, mockMessageRepo)

			group, err := chatService.GetGroupByID(tt.groupID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, group)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, group)
				assert.Equal(t, tt.groupID, group.ID)
			}

			mockGroupRepo.AssertExpectations(t)
		})
	}
}

func TestChatService_GetGroupsByMember(t *testing.T) {
	mockGroupRepo := new(chatMocks.MockGroupRepository)
	mockMessageRepo := new(chatMocks.MockMessageRepository)

	expectedGroups := []*chatModel.Group{
		{
			Base: model.Base{
				ID:        "group1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "Group 1",
			CreatorID: "user123",
			MemberIDs: []string{"user123", "member1"},
			Slug:      "group-1-abc123",
		},
		{
			Base: model.Base{
				ID:        "group2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "Group 2",
			CreatorID: "other-user",
			MemberIDs: []string{"user123", "member2"},
			Slug:      "group-2-def456",
		},
	}

	mockGroupRepo.On("GetByMemberID", "user123").Return(expectedGroups, nil)

	chatService := chat.NewChatService(mockGroupRepo, mockMessageRepo)

	groups, err := chatService.GetGroupsByMember("user123")

	assert.NoError(t, err)
	assert.Len(t, groups, 2)
	assert.Equal(t, "group1", groups[0].ID)
	assert.Equal(t, "group2", groups[1].ID)

	mockGroupRepo.AssertExpectations(t)
}

func TestChatService_AddMemberToGroup(t *testing.T) {
	tests := []struct {
		name        string
		groupID     string
		memberID    string
		requesterID string
		expectedErr error
		mockSetup   func(*chatMocks.MockGroupRepository)
	}{
		{
			name:        "successful member addition by creator",
			groupID:     "group123",
			memberID:    "newmember",
			requesterID: "creator123",
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "group123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "creator123",
					MemberIDs: []string{"creator123", "member1"},
					Slug:      "test-group-abc123",
				}
				mockRepo.On("GetByID", "group123").Return(group, nil)
				mockRepo.On("AddMember", "group123", "newmember").Return(nil)
			},
		},
		{
			name:        "successful member addition by existing member",
			groupID:     "group123",
			memberID:    "newmember",
			requesterID: "member1",
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "group123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "creator123",
					MemberIDs: []string{"creator123", "member1"},
					Slug:      "test-group-abc123",
				}
				mockRepo.On("GetByID", "group123").Return(group, nil)
				mockRepo.On("AddMember", "group123", "newmember").Return(nil)
			},
		},
		{
			name:        "member already exists",
			groupID:     "group123",
			memberID:    "member1",
			requesterID: "creator123",
			expectedErr: errors.New("member already exists in group"),
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "group123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "creator123",
					MemberIDs: []string{"creator123", "member1"}, // member1 is already in the group
					Slug:      "test-group-abc123",
				}
				mockRepo.On("GetByID", "group123").Return(group, nil)
				mockRepo.On("AddMember", "group123", "member1").Return(errors.New("member already exists in group"))
			},
		},
		{
			name:        "not authorized",
			groupID:     "group123",
			memberID:    "newmember",
			requesterID: "outsider",
			expectedErr: errors.New("only group members can add new members"),
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "group123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "creator123",
					MemberIDs: []string{"creator123", "member1"},
					Slug:      "test-group-abc123",
				}
				mockRepo.On("GetByID", "group123").Return(group, nil)
			},
		},
		{
			name:        "group not found",
			groupID:     "nonexistent",
			memberID:    "newmember",
			requesterID: "creator123",
			expectedErr: errors.New("group not found"),
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				mockRepo.On("GetByID", "nonexistent").Return(nil, errors.New("group not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupRepo := new(chatMocks.MockGroupRepository)
			mockMessageRepo := new(chatMocks.MockMessageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockGroupRepo)
			}

			chatService := chat.NewChatService(mockGroupRepo, mockMessageRepo)

			err := chatService.AddMemberToGroup(tt.groupID, tt.memberID, tt.requesterID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockGroupRepo.AssertExpectations(t)
		})
	}
}

func TestChatService_RemoveMemberFromGroup(t *testing.T) {
	tests := []struct {
		name        string
		groupID     string
		memberID    string
		requesterID string
		expectedErr error
		mockSetup   func(*chatMocks.MockGroupRepository)
	}{
		{
			name:        "successful member removal by creator",
			groupID:     "group123",
			memberID:    "member1",
			requesterID: "creator123",
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "group123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "creator123",
					MemberIDs: []string{"creator123", "member1", "member2"},
					Slug:      "test-group-abc123",
				}
				mockRepo.On("GetByID", "group123").Return(group, nil)
				mockRepo.On("RemoveMember", "group123", "member1").Return(nil)
			},
		},
		{
			name:        "successful self-removal",
			groupID:     "group123",
			memberID:    "member1",
			requesterID: "member1",
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "group123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "creator123",
					MemberIDs: []string{"creator123", "member1", "member2"},
					Slug:      "test-group-abc123",
				}
				mockRepo.On("GetByID", "group123").Return(group, nil)
				mockRepo.On("RemoveMember", "group123", "member1").Return(nil)
			},
		},
		{
			name:        "not authorized",
			groupID:     "group123",
			memberID:    "member1",
			requesterID: "outsider",
			expectedErr: errors.New("only group creator or the member themselves can remove a member"),
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "group123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "creator123",
					MemberIDs: []string{"creator123", "member1", "member2"},
					Slug:      "test-group-abc123",
				}
				mockRepo.On("GetByID", "group123").Return(group, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupRepo := new(chatMocks.MockGroupRepository)
			mockMessageRepo := new(chatMocks.MockMessageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockGroupRepo)
			}

			chatService := chat.NewChatService(mockGroupRepo, mockMessageRepo)

			err := chatService.RemoveMemberFromGroup(tt.groupID, tt.memberID, tt.requesterID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockGroupRepo.AssertExpectations(t)
		})
	}
}

func TestChatService_SendMessage(t *testing.T) {
	tests := []struct {
		name        string
		senderID    string
		groupID     string
		content     string
		expectedErr error
		mockSetup   func(*chatMocks.MockGroupRepository, *chatMocks.MockMessageRepository)
	}{
		{
			name:     "successful message send",
			senderID: "user123",
			groupID:  "group123",
			content:  "Hello, world!",
			mockSetup: func(mockGroupRepo *chatMocks.MockGroupRepository, mockMessageRepo *chatMocks.MockMessageRepository) {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "group123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "creator123",
					MemberIDs: []string{"creator123", "user123", "member1"},
					Slug:      "test-group-abc123",
				}
				mockGroupRepo.On("GetByID", "group123").Return(group, nil)
				mockMessageRepo.On("Create", mock.MatchedBy(func(m *chatModel.Message) bool {
					return m.SenderID == "user123" &&
						m.GroupID == "group123" &&
						m.Content == "Hello, world!"
				})).Return(nil)
			},
		},
		{
			name:        "sender not a member",
			senderID:    "outsider",
			groupID:     "group123",
			content:     "Hello, world!",
			expectedErr: errors.New("sender is not a member of this group"),
			mockSetup: func(mockGroupRepo *chatMocks.MockGroupRepository, mockMessageRepo *chatMocks.MockMessageRepository) {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "group123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "creator123",
					MemberIDs: []string{"creator123", "member1"},
					Slug:      "test-group-abc123",
				}
				mockGroupRepo.On("GetByID", "group123").Return(group, nil)
			},
		},
		{
			name:        "group not found",
			senderID:    "user123",
			groupID:     "nonexistent",
			content:     "Hello, world!",
			expectedErr: errors.New("group not found"),
			mockSetup: func(mockGroupRepo *chatMocks.MockGroupRepository, mockMessageRepo *chatMocks.MockMessageRepository) {
				mockGroupRepo.On("GetByID", "nonexistent").Return(nil, errors.New("group not found"))
			},
		},
		{
			name:        "message content too long",
			senderID:    "user123",
			groupID:     "group123",
			content:     string(make([]byte, 1001)), // 1001 characters
			expectedErr: errors.New("message content must be at most 1000 characters"),
			mockSetup: func(mockGroupRepo *chatMocks.MockGroupRepository, mockMessageRepo *chatMocks.MockMessageRepository) {
				// No repository calls should be made since validation fails first
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupRepo := new(chatMocks.MockGroupRepository)
			mockMessageRepo := new(chatMocks.MockMessageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockGroupRepo, mockMessageRepo)
			}

			chatService := chat.NewChatService(mockGroupRepo, mockMessageRepo)

			message, err := chatService.SendMessage(tt.senderID, tt.groupID, tt.content)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
				assert.Nil(t, message)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, message)
				assert.Equal(t, tt.senderID, message.SenderID)
				assert.Equal(t, tt.groupID, message.GroupID)
				assert.Equal(t, tt.content, message.Content)
			}

			mockGroupRepo.AssertExpectations(t)
			mockMessageRepo.AssertExpectations(t)
		})
	}
}

func TestChatService_UpdateGroup(t *testing.T) {
	tests := []struct {
		name        string
		group       *chatModel.Group
		expectedErr error
		mockSetup   func(*chatMocks.MockGroupRepository)
	}{
		{
			name: "successful group update",
			group: &chatModel.Group{
				Base: model.Base{
					ID:        "group123",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:      "Updated Group",
				CreatorID: "creator123",
				MemberIDs: []string{"creator123", "member1"},
				Slug:      "updated-group-abc123",
				Image:     "updated.jpg",
			},
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				mockRepo.On("Update", mock.MatchedBy(func(g *chatModel.Group) bool {
					return g.Name == "Updated Group" && g.Image == "updated.jpg"
				})).Return(nil)
			},
		},
		{
			name: "group name too long",
			group: &chatModel.Group{
				Base: model.Base{
					ID:        "group123",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:      "This is a very long group name that exceeds the 20 character limit",
				CreatorID: "creator123",
				MemberIDs: []string{"creator123", "member1"},
				Slug:      "test-group-abc123",
				Image:     "test.jpg",
			},
			expectedErr: errors.New("group name must be at most 20 characters"),
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				// No repository call expected due to validation failure
			},
		},
		{
			name: "repository error",
			group: &chatModel.Group{
				Base: model.Base{
					ID:        "group123",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:      "Test Group",
				CreatorID: "creator123",
				MemberIDs: []string{"creator123", "member1"},
				Slug:      "test-group-abc123",
				Image:     "test.jpg",
			},
			expectedErr: errors.New("repository error"),
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				mockRepo.On("Update", mock.Anything).Return(errors.New("repository error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupRepo := new(chatMocks.MockGroupRepository)
			mockMessageRepo := new(chatMocks.MockMessageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockGroupRepo)
			}

			chatService := chat.NewChatService(mockGroupRepo, mockMessageRepo)

			err := chatService.UpdateGroup(tt.group)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockGroupRepo.AssertExpectations(t)
		})
	}
}

func TestChatService_DeleteGroup(t *testing.T) {
	tests := []struct {
		name        string
		groupID     string
		requesterID string
		expectedErr error
		mockSetup   func(*chatMocks.MockGroupRepository)
	}{
		{
			name:        "successful group deletion by creator",
			groupID:     "group123",
			requesterID: "creator123",
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "group123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "creator123",
					MemberIDs: []string{"creator123", "member1"},
					Slug:      "test-group-abc123",
				}
				mockRepo.On("GetByID", "group123").Return(group, nil)
				mockRepo.On("Delete", "group123").Return(nil)
			},
		},
		{
			name:        "not authorized",
			groupID:     "group123",
			requesterID: "member1",
			expectedErr: errors.New("only the group creator can delete the group"),
			mockSetup: func(mockRepo *chatMocks.MockGroupRepository) {
				group := &chatModel.Group{
					Base: model.Base{
						ID:        "group123",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:      "Test Group",
					CreatorID: "creator123",
					MemberIDs: []string{"creator123", "member1"},
					Slug:      "test-group-abc123",
				}
				mockRepo.On("GetByID", "group123").Return(group, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGroupRepo := new(chatMocks.MockGroupRepository)
			mockMessageRepo := new(chatMocks.MockMessageRepository)

			if tt.mockSetup != nil {
				tt.mockSetup(mockGroupRepo)
			}

			chatService := chat.NewChatService(mockGroupRepo, mockMessageRepo)

			err := chatService.DeleteGroup(tt.groupID, tt.requesterID)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mockGroupRepo.AssertExpectations(t)
		})
	}
}

func TestChatService_GetMessagesByGroup(t *testing.T) {
	mockGroupRepo := new(chatMocks.MockGroupRepository)
	mockMessageRepo := new(chatMocks.MockMessageRepository)

	expectedMessages := []*chatModel.Message{
		{
			Base: model.Base{
				ID:        "msg1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			SenderID: "user1",
			Content:  "Hello!",
			GroupID:  "group123",
		},
		{
			Base: model.Base{
				ID:        "msg2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			SenderID: "user2",
			Content:  "Hi there!",
			GroupID:  "group123",
		},
	}

	mockMessageRepo.On("GetByGroupID", "group123", 20, 0).Return(expectedMessages, nil)

	chatService := chat.NewChatService(mockGroupRepo, mockMessageRepo)

	messages, err := chatService.GetMessagesByGroup("group123", 20, 0)

	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "msg1", messages[0].ID)
	assert.Equal(t, "msg2", messages[1].ID)
	assert.Equal(t, "Hello!", messages[0].Content)
	assert.Equal(t, "Hi there!", messages[1].Content)

	mockMessageRepo.AssertExpectations(t)
}

func TestChatService_SearchGroupsByName(t *testing.T) {
	mockGroupRepo := new(chatMocks.MockGroupRepository)
	mockMessageRepo := new(chatMocks.MockMessageRepository)

	expectedGroups := []*chatModel.Group{
		{
			Base: model.Base{
				ID:        "group1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Name:      "Test Group",
			CreatorID: "user123",
			MemberIDs: []string{"user123", "member1"},
			Slug:      "test-group-abc123",
		},
	}

	mockGroupRepo.On("SearchByName", "test", 10, 0).Return(expectedGroups, nil)

	chatService := chat.NewChatService(mockGroupRepo, mockMessageRepo)

	groups, err := chatService.SearchGroupsByName("test", 10, 0)

	assert.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.Equal(t, "group1", groups[0].ID)
	assert.Equal(t, "Test Group", groups[0].Name)

	mockGroupRepo.AssertExpectations(t)
}
